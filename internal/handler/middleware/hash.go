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
				h.ServeHTTP(w, r)
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

			rw := &hashResponseWriter{
				ResponseWriter: w,
				body:           bytes.NewBuffer(nil),
			}

			h.ServeHTTP(rw, r)

			if rw.body.Len() > 0 {
				responseHash, err := util.HashByKey(rw.body.Bytes(), key)
				if err != nil {
					logger.Log.Warn("Failed to hash body", zap.Error(err))
					http.Error(w, "", http.StatusInternalServerError)
					return
				}
				rw.Header().Set("HashSHA256", responseHash)
			}

			rw.ResponseWriter.Write(rw.body.Bytes())
		})
	}
}

type hashResponseWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (w *hashResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *hashResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *hashResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}
