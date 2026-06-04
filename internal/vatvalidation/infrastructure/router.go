package infrastructure

import (
	"log/slog"
	"net/http"

	"codeberg.org/simplicer/vatguard/internal/vatvalidation/application"
)

// NewRouter creates the HTTP router with all routes configured.
func NewRouter(vies application.VIESChecker, log *slog.Logger, version string) http.Handler {
	handler := NewHTTPHandler(vies, log, version)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/vat/{number}", handler.HandleVAT)
	mux.HandleFunc("GET /v1/vies/{number}", handler.HandleVIES)
	mux.HandleFunc("GET /health", handler.HandleHealth)
	mux.HandleFunc("GET /version", handler.HandleVersion)

	return loggingMiddleware(mux, log)
}

func loggingMiddleware(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("request received",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
		)
		next.ServeHTTP(w, r)
	})
}
