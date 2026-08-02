package handlers

import (
	"net"
	"net/http"
	"time"

	"url-shortener/database"
)

func RateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr // Fallback to raw address
		}

		key := "rate_limit:" + ip
		pipe := database.RedisClient.Pipeline()
		incr := pipe.Incr(database.Ctx, key)
		pipe.Expire(database.Ctx, key, time.Minute)
		_, err = pipe.Exec(database.Ctx)

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if incr.Val() > 5 {
			http.Error(w, "429 Too Many Requests: Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
