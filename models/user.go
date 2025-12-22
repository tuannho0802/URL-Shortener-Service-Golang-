package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`          // Save hash not plain text
	Links    []Link `gorm:"foreignKey:UserID"` // 1 user - many links
}
