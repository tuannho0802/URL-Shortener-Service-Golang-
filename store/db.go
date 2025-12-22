package store

import (
	"time"

	"github.com/glebarez/sqlite"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/models"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
    var err error
    // Open connection
    DB, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    if err != nil {
        panic("Failed to connect to database: " + err.Error())
    }

    // Activate SQLite's WAL mode
    // Read and write same time and not block
    sqlDB, err := DB.DB()
    if err == nil {
        DB.Exec("PRAGMA journal_mode=WAL;")
        DB.Exec("PRAGMA synchronous=NORMAL;")
        
        // Config connection pool
       
        sqlDB.SetMaxOpenConns(100)           // Limited 100 connection
        sqlDB.SetMaxIdleConns(10)            // keep alive 10 connection
        sqlDB.SetConnMaxLifetime(time.Hour)  // set max lifetime connection
    }

    // Auto migrate
    DB.AutoMigrate(&models.User{}, &models.Link{})
}