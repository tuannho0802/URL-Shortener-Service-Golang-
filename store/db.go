package store

import (
	"log"
	"os"
	"time"

	"github.com/tuannho0802/URL-Shortener-Service-Golang-/models"
	"golang.org/x/crypto/bcrypt" // Import bcrypt to hash password
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=password dbname=shortener port=5432 sslmode=disable"
	}

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Database connection failed: ", err)
	}

	// Optimize connection pool
	sqlDB, err := DB.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	// Automigrate schemas
	err = DB.AutoMigrate(&models.User{}, &models.Link{})
	if err != nil {
		log.Fatal("❌ Migration failed: ", err)
	}

	// Seed: Automatically create or update Admin account
	seedAdmin()
}

func seedAdmin() {
	adminUsername := "admin"
	adminPassword := "admin"

	// Check if admin already exists
	var user models.User
	result := DB.Where("username = ?", adminUsername).First(&user)

	// Hash password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)

	if result.Error == gorm.ErrRecordNotFound {
		// Case 1: Create new Admin if not exists
		newAdmin := models.User{
			Username: adminUsername,
			Password: string(hashedPassword),
			Role:     "admin",
		}
		if err := DB.Create(&newAdmin).Error; err != nil {
			log.Println("⚠️ Failed to create seed admin:", err)
		} else {
			log.Println("=== ✅ Seeded default admin (admin/admin) ===")
		}
	} else {
		// Case 2: Ensure existing 'admin' user always has 'admin' role
		DB.Model(&user).Update("role", "admin")
		log.Println("=== ✅ Admin role verified for existing user: admin ===")
	}
}