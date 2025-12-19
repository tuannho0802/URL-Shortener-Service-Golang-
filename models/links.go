package models

import "time"

// link table in database
type Link struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	OriginalURL string    `gorm:"not null" json:"original_url"`
	ShortCode   string    `gorm:"unique; not null" json:"short_code"` // unique short code
	ClickCount  int       `gorm:"default:0" json:"click_count"`       // click count
	CreatedAt   time.Time `json:"created_at"`
}
