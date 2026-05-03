package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/AnxVit/metrics-server/internal/logger"
	"github.com/AnxVit/metrics-server/internal/util"
	"go.uber.org/zap"
)

func HashMiddleware(key string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				h.ServeHTTP(w, r)
				return
			}

			hash := r.Header.Get("HashSHA256")
			if hash == "" {
				logger.Log.Info("Missing hash header")
				http.Error(w, "Missing hash header", http.StatusBadRequest)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log.Warn("Failed to read body", zap.Error(err))
				http.Error(w, "Failed to read body", http.StatusUnprocessableEntity)
				return
			}
			defer r.Body.Close()

			if !util.VerifyHash(body, hash, key) {
				logger.Log.Info("Not verified hash", zap.Error(err))
				http.Error(w, "Not verified hash", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(body))

			h.ServeHTTP(w, r)
		})
	}
}
