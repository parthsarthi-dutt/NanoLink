package models

import "time"

type URL struct {
	ID          int        `json:"id"`
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	Clicks      int        `json:"clicks"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ShortenRequest struct {
	OriginalURL string     `json:"original_url"`
	CustomAlias string     `json:"custom_alias,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type ShortenResponse struct {
	ShortURL  string     `json:"short_url"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
