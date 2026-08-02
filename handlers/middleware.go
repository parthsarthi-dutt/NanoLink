package handlers

import (
	"net"
	"net/http"
	"time"

	"url-shortener/database"
)

// RateLimiter limits requests to 50 per minute per IP using Redis
func RateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract IP address
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr // Fallback to raw address
		}

		key := "rate_limit:" + ip

		// Atomic increment in Redis
		pipe := database.RedisClient.Pipeline()
		incr := pipe.Incr(database.Ctx, key)
		pipe.Expire(database.Ctx, key, time.Minute)
		_, err = pipe.Exec(database.Ctx)

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Check limit
		if incr.Val() > 5 {
			http.Error(w, "429 Too Many Requests: Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Proceed to handler
		next.ServeHTTP(w, r)
	})
}
