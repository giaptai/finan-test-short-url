package model

import "time"

type URL struct {
	ID          int64     `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	Clicks      int64     `json:"clicks"`
	CreatedAt   time.Time `json:"create_at"`
}
