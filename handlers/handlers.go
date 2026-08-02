package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"url-shortener/database"
	"url-shortener/models"
	"url-shortener/utils"
)

func getBaseURL() string {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080/"
	}
	return baseURL
}

func ShortenURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.OriginalURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if _, err := url.ParseRequestURI(req.OriginalURL); err != nil {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	shortCode := req.CustomAlias
	if shortCode == "" {
		for {
			shortCode = utils.GenerateShortCode(6)
			
			var exists bool
			err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM urls WHERE short_code=$1)", shortCode).Scan(&exists)
			if err != nil {
				log.Println("Database error checking collision:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			
			if !exists {
				break // Code is unique
			}
		}
	} else {
		var exists bool
		err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM urls WHERE short_code=$1)", shortCode).Scan(&exists)
		if err != nil {
			log.Println("Database error checking custom alias:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if exists {
			http.Error(w, "Custom alias already in use", http.StatusConflict)
			return
		}
	}

	_, err := database.DB.Exec(
		"INSERT INTO urls (short_code, original_url, expires_at) VALUES ($1, $2, $3)",
		shortCode, req.OriginalURL, req.ExpiresAt,
	)
	if err != nil {
		log.Println("Database error inserting URL:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := models.ShortenResponse{
		ShortURL:  getBaseURL() + shortCode,
		ExpiresAt: req.ExpiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func RedirectURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Strip the leading '/' to get the short code
	shortCode := r.URL.Path[1:]
	if shortCode == "" {
		http.Error(w, "Short code required", http.StatusBadRequest)
		return
	}

	var originalURL string
	var expiresAt sql.NullTime

	cachedURL, err := database.RedisClient.Get(database.Ctx, "url:"+shortCode).Result()
	if err == nil && cachedURL != "" {
		go func(code string) {
			_, err := database.DB.Exec("UPDATE urls SET clicks = clicks + 1 WHERE short_code = $1", code)
			if err != nil {
				log.Println("Error updating click count:", err)
			}
		}(shortCode)
		http.Redirect(w, r, cachedURL, http.StatusFound)
		return
	}

	err = database.DB.QueryRow(
		"SELECT original_url, expires_at FROM urls WHERE short_code=$1", shortCode,
	).Scan(&originalURL, &expiresAt)

	if err == sql.ErrNoRows {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Println("Database error fetching URL:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		http.Error(w, "URL has expired", http.StatusGone)
		return
	}

	database.RedisClient.Set(database.Ctx, "url:"+shortCode, originalURL, time.Hour)

	go func(code string) {
		_, err := database.DB.Exec("UPDATE urls SET clicks = clicks + 1 WHERE short_code = $1", code)
		if err != nil {
			log.Println("Error updating click count:", err)
		}
	}(shortCode)

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shortCode := r.URL.Query().Get("code")
	if shortCode == "" {
		http.Error(w, "Code parameter is required", http.StatusBadRequest)
		return
	}

	var urlData models.URL
	err := database.DB.QueryRow(
		"SELECT id, short_code, original_url, clicks, expires_at, created_at FROM urls WHERE short_code=$1",
		shortCode,
	).Scan(&urlData.ID, &urlData.ShortCode, &urlData.OriginalURL, &urlData.Clicks, &urlData.ExpiresAt, &urlData.CreatedAt)

	if err == sql.ErrNoRows {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Println("Database error fetching stats:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(urlData)
}

func GetAllURLsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := database.DB.Query("SELECT id, short_code, original_url, clicks, expires_at, created_at FROM urls ORDER BY created_at DESC")
	if err != nil {
		log.Println("Database error fetching all URLs:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var urls []models.URL
	for rows.Next() {
		var u models.URL
		if err := rows.Scan(&u.ID, &u.ShortCode, &u.OriginalURL, &u.Clicks, &u.ExpiresAt, &u.CreatedAt); err != nil {
			log.Println("Database error scanning URL:", err)
			continue
		}
		urls = append(urls, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(urls)
}

func UpdateURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shortCode := r.URL.Query().Get("code")
	if shortCode == "" {
		http.Error(w, "Code parameter is required", http.StatusBadRequest)
		return
	}

	var req models.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.OriginalURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	res, err := database.DB.Exec(
		"UPDATE urls SET original_url=$1, expires_at=$2 WHERE short_code=$3",
		req.OriginalURL, req.ExpiresAt, shortCode,
	)
	if err != nil {
		log.Println("Database error updating URL:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "URL updated successfully"}`))
}

func DeleteURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shortCode := r.URL.Query().Get("code")
	if shortCode == "" {
		http.Error(w, "Code parameter is required", http.StatusBadRequest)
		return
	}

	res, err := database.DB.Exec("DELETE FROM urls WHERE short_code=$1", shortCode)
	if err != nil {
		log.Println("Database error deleting URL:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
