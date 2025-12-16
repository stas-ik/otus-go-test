package internalhttp

import (
	"fmt"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.written += n
	return n, err
}

// loggingMiddleware логирует HTTP запросы.
func loggingMiddleware(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			clientIP := r.RemoteAddr
			if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
				clientIP = forwardedFor
			}

			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     0,
				written:        0,
			}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			latencyMs := duration.Milliseconds()

			userAgent := r.UserAgent()
			if userAgent == "" {
				userAgent = "-"
			}

			// 66.249.65.3 [25/Feb/2020:19:11:24 +0600] GET /hello?q=1 HTTP/1.1 200 30 "Mozilla/5.0"
			logLine := fmt.Sprintf(
				`%s [%s] %s %s %s %d %d %dms "%s"`,
				clientIP,
				start.Format("02/Jan/2006:15:04:05 -0700"),
				r.Method,
				r.RequestURI,
				r.Proto,
				rw.statusCode,
				rw.written,
				latencyMs,
				userAgent,
			)

			logger.Info(logLine)
		})
	}
}
