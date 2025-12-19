package handlers

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/models"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/store"
)

// Generate random code
func GenerateShortCode(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// Create short link
func CreateShortLink(c *gin.Context) {

	var input struct {
		OriginalURL string `json:"original_url"`
		CustomAlias string `json:"custom_alias"`
		// Add expires_in_days
		ExpiresInDays int `json:"expires_in_days"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// Calculate the expiration date
	var expiredAt *time.Time
	if input.ExpiresInDays > 0 {
		t := time.Now().AddDate(0, 0, input.ExpiresInDays)
		expiredAt = &t
	} else {
		// Default 1 day
		t := time.Now().Add(24 * time.Hour)
		expiredAt = &t
	}

	var code string
	if input.CustomAlias != "" {
		// Check if Allias exist
		var existing models.Link
		if err := store.DB.Where("short_code = ?", input.CustomAlias).First(&existing).Error; err == nil {
			c.JSON(400, gin.H{"error": "Alias này đã được sử dụng"})
			return
		}
		code = input.CustomAlias
	} else {
		code = GenerateShortCode(6) // Generate 6 random code
	}

	// Save on db
	newLink := models.Link{
		OriginalURL: input.OriginalURL,
		ShortCode:   code,
		ClickCount:  0,
		ExpiredAt:   expiredAt, // Store db
	}
	store.DB.Create(&newLink)

	// notify
	go NotifyDataChange()

	c.JSON(200, gin.H{"short_url": "http://localhost:8080/" + code})
}

// 2. API Redirect
func RedirectLink(c *gin.Context) {
	code := c.Param("code")
	var link models.Link

	// Find link in ShortCode
	if err := store.DB.Where("short_code = ?", code).First(&link).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	// Check if link is expired
	if link.ExpiredAt != nil && time.Now().After(*link.ExpiredAt) {
		c.JSON(410, gin.H{"error": "Link này đã hết hạn và không còn khả dụng"})
		return
	}

	// Update click if link is not expired
	store.DB.Model(&link).Update("click_count", link.ClickCount+1)

	// notify data
	go NotifyDataChange()

	// Redirect main page
	c.Redirect(http.StatusFound, link.OriginalURL)
}

// 3. API get list
func GetAllLinks(c *gin.Context) {
	var links []models.Link
	store.DB.Find(&links)
	c.JSON(http.StatusOK, links)
}
