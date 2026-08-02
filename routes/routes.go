package routes

import (
	"net/http"
	"strings"

	"url-shortener/handlers"
)

func RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/shorten", handlers.ShortenURLHandler)
	mux.HandleFunc("/api/stats", handlers.StatsHandler)
	mux.HandleFunc("/api/urls", handlers.GetAllURLsHandler)
	mux.HandleFunc("/api/urls/update", handlers.UpdateURLHandler)
	mux.HandleFunc("/api/urls/delete", handlers.DeleteURLHandler)
	mux.HandleFunc("/health", handlers.HealthCheckHandler)

	return handlers.RateLimiter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/health" {
			mux.ServeHTTP(w, r)
			return
		}

		if r.URL.Path != "/" && !strings.Contains(r.URL.Path[1:], "/") {
			handlers.RedirectURLHandler(w, r)
			return
		}

		http.NotFound(w, r)
	}))
}
