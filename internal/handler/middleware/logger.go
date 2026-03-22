package middleware

import (
	"net/http"
	"time"

	"github.com/AnxVit/metrics-server/internal/logger"
	"go.uber.org/zap"
)

type responseDate struct {
	status int
	size   int
}

type LogResponseWriter struct {
	http.ResponseWriter
	responseDate *responseDate
}

func (l *LogResponseWriter) Write(b []byte) (int, error) {
	size, err := l.ResponseWriter.Write(b)
	l.responseDate.size = size
	return size, err
}

func (l *LogResponseWriter) WriteHeader(statusCode int) {
	l.ResponseWriter.WriteHeader(statusCode)
	l.responseDate.status = statusCode
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		responseDate := &responseDate{
			status: 0,
			size:   0,
		}
		lw := LogResponseWriter{
			ResponseWriter: w,
			responseDate:   responseDate,
		}
		next.ServeHTTP(&lw, r)

		duration := time.Since(startedAt)

		logger.Log.Info("",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", responseDate.status),
			zap.Duration("duration", duration),
			zap.Int("size", responseDate.size),
		)
	})
}
