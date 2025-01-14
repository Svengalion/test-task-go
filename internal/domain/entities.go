package domain

import "time"

type Song struct {
	ID          int64     `json:"id"`
	Group       string    `json:"group"`
	Title       string    `json:"song"`
	ReleaseDate time.Time `json:"releaseDate"`
	Text        string    `json:"text"`
	Link        string    `json:"link"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type SongFilter struct {
	Group    *string
	Title    *string
	Page     int
	PageSize int
}
