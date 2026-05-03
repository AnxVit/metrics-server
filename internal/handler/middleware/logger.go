package middleware

import (
	"net/http"
	"time"

	"github.com/AnxVit/metrics-server/internal/logger"
	"go.uber.org/zap"
)

type responseData struct {
	status int
	size   int
}

type LogResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (l *LogResponseWriter) Write(b []byte) (int, error) {
	size, err := l.ResponseWriter.Write(b)
	l.responseData.size = size
	return size, err
}

func (l *LogResponseWriter) WriteHeader(statusCode int) {
	l.ResponseWriter.WriteHeader(statusCode)
	l.responseData.status = statusCode
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := LogResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		next.ServeHTTP(&lw, r)

		duration := time.Since(startedAt)

		logger.Log.Info("",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", responseData.status),
			zap.Duration("duration", duration),
			zap.Int("size", responseData.size),
		)
	})
}
