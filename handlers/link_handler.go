package handlers

import (
	"math/rand"
	"net/http"
	"strings"
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

// API Redirect
func RedirectLink(c *gin.Context) {
	var link models.Link
	code := c.Param("code")

	// check link exist
	if err := store.DB.Where("short_code = ?", code).First(&link).Error; err != nil {
		c.JSON(404, gin.H{"error": "Link không tồn tại"})
		return
	}

	// Check Browser from user agent
	ua := c.GetHeader("User-Agent")
	browser := "Unknown"
	os := "Unknown"

	// Logic check
	if strings.Contains(ua, "Firefox") {
		browser = "Firefox"
	} else if strings.Contains(ua, "Chrome") && !strings.Contains(ua, "Edg") {
		browser = "Chrome"
	} else if strings.Contains(ua, "Safari") && !strings.Contains(ua, "Chrome") {
		browser = "Safari"
	} else if strings.Contains(ua, "Edg") {
		browser = "Edge"
	}

	// Check OS
	if strings.Contains(ua, "Windows") {
		os = "Windows"
	} else if strings.Contains(ua, "Macintosh") {
		os = "MacOS"
	} else if strings.Contains(ua, "Android") {
		os = "Android"
	} else if strings.Contains(ua, "iPhone") {
		os = "iOS"
	}

	// Update DB
	store.DB.Model(&link).Updates(map[string]interface{}{
		"click_count":  link.ClickCount + 1,
		"last_browser": browser,
		"last_os":      os,
	})

	go NotifyDataChange()

	c.Redirect(302, link.OriginalURL)
}

// API get all list
func GetAllLinks(c *gin.Context) {
	var links []models.Link
	sortParam := c.DefaultQuery("sort", "created_at_desc")

	query := store.DB
	switch sortParam {
	case "abc":
		query = query.Order("short_code ASC")
	case "clicks":
		query = query.Order("click_count DESC")
	case "oldest":
		query = query.Order("created_at ASC")
	default: // newest
		query = query.Order("created_at DESC")
	}

	query.Find(&links)
	c.JSON(http.StatusOK, links)
}

// Delete link
func DeleteLink(c *gin.Context) {
	id := c.Param("id")

	// Remove in db
	if err := store.DB.Delete(&models.Link{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa liên kết"})
		return
	}

	// send notify
	go NotifyDataChange()

	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa liên kết thành công"})
}

// Edit link
func UpdateLink(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		OriginalURL string `json:"original_url"`
	}

	// Check data input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// Update new URL
	if err := store.DB.Model(&models.Link{}).Where("id = ?", id).Update("original_url", input.OriginalURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể cập nhật liên kết"})
		return
	}

	// Notify
	go NotifyDataChange()

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật thành công"})
}
