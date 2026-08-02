package routes

import (
	"net/http"
	"strings"

	"url-shortener/handlers"
)

func RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register specific API routes
	mux.HandleFunc("/api/shorten", handlers.ShortenURLHandler)
	mux.HandleFunc("/api/stats", handlers.StatsHandler)
	mux.HandleFunc("/api/urls", handlers.GetAllURLsHandler)
	mux.HandleFunc("/api/urls/update", handlers.UpdateURLHandler)
	mux.HandleFunc("/api/urls/delete", handlers.DeleteURLHandler)
	mux.HandleFunc("/health", handlers.HealthCheckHandler)

	// Wrap the mux to handle dynamic short code routes at the root level, and wrap everything in the RateLimiter
	return handlers.RateLimiter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Pass API routes and health check to the standard mux
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/health" {
			mux.ServeHTTP(w, r)
			return
		}

		// Handle short code redirection (e.g., /ab12cd)
		// Ignore root path "/" or paths with multiple segments like "/foo/bar"
		if r.URL.Path != "/" && !strings.Contains(r.URL.Path[1:], "/") {
			handlers.RedirectURLHandler(w, r)
			return
		}

		http.NotFound(w, r)
	}))
}
