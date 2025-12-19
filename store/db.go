package store

import (
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	// Use SQLite and the file is name "test.db"
	DB, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}

	// Auto migrate base on struct link
	DB.AutoMigrate(&models.Link{})
}
