package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username         string `gorm:"unique;not null"`
	Password         string `gorm:"not null"` // Save hash not plain text
	Role             string `gorm:"default:user"`
	Links            []Link `gorm:"foreignKey:UserID"` // 1 user - many links
	ResetToken       string `gorm:"index"`
	ResetTokenExpiry time.Time
}
