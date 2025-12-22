package models

import "time"

// link table in database
type Link struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	OriginalURL string     `gorm:"not null" json:"original_url"`
	ShortCode   string     `gorm:"uniqueIndex;not null" json:"short_code"` // unique short code
	ClickCount  int        `gorm:"default:0" json:"click_count"`           // click count
	UserID      uint       `json:"user_id"`                                // user id
	LastBrowser string     `json:"last_browser"`
	LastOS      string     `json:"last_os"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiredAt   *time.Time `gorm:"index" json:"expired_at"` // expired code

}
