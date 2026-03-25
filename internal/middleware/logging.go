package middleware

import (
	"net/http"
	"time"

	"github.com/alxaxenov/ya-gophermart/internal/logger"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// WithLogging middleware для логирования запросов. Логирует метод, uri, время выполнения, код ответа, размер тела ответа.
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method
		rw := &loggingResponseWriter{ResponseWriter: w, responseData: &responseData{}}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		logger.Logger.Infoln(
			"method:", method,
			"uri:", uri,
			"duration:", duration,
			"response_status:", rw.responseData.status,
			"response_size:", rw.responseData.size,
		)
	})
}
