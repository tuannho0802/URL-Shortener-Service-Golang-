package handlers

import (
	"fmt"
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
		OriginalURL   string `json:"original_url"`
		CustomAlias   string `json:"custom_alias"`
		DurationType  string `json:"duration_type"` // "minutes", "hours", "days", "infinite"
		DurationValue int    `json:"duration_value"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// Debug
	fmt.Printf("DEBUG: Type=%s, Value=%d\n", input.DurationType, input.DurationValue)

	// Logic check
	var expiredAt *time.Time

	// Check duration if empty default 1 day
	if input.DurationType == "" {
		input.DurationType = "days"
		input.DurationValue = 1
	}

	now := time.Now()
	switch input.DurationType {
	case "minutes":
		t := now.Add(time.Duration(input.DurationValue) * time.Minute)
		expiredAt = &t
	case "hours":
		t := now.Add(time.Duration(input.DurationValue) * time.Hour)
		expiredAt = &t
	case "days":
		t := now.AddDate(0, 0, input.DurationValue)
		expiredAt = &t
	case "infinite":
		expiredAt = nil // infinite
	default:
		// Return default if not match data input
		t := now.AddDate(0, 0, 1)
		expiredAt = &t
	}

	var code string
	if input.CustomAlias != "" {
		var existing models.Link
		if err := store.DB.Where("short_code = ?", input.CustomAlias).First(&existing).Error; err == nil {
			c.JSON(400, gin.H{"error": "Alias này đã được sử dụng"})
			return
		}
		code = input.CustomAlias
	} else {
		code = GenerateShortCode(6)
	}

	newLink := models.Link{
		OriginalURL: input.OriginalURL,
		ShortCode:   code,
		ClickCount:  0,
		ExpiredAt:   expiredAt,
	}

	if err := store.DB.Create(&newLink).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lưu vào database"})
		return
	}

	go NotifyDataChange()

	c.JSON(200, gin.H{"short_url": "http://localhost:8080/" + code})
}

// API Redirect
func RedirectLink(c *gin.Context) {
	var link models.Link
	code := c.Param("code")

	fmt.Printf("DEBUG: Đang truy cập mã code [%s]\n", code)

	// If code is system file or null skip
	if code == "" || code == "/" || code == "index.html" {

		return
	}
	// check link exist
	if err := store.DB.Where("short_code = ?", code).First(&link).Error; err != nil {
		// Instead of render expired.html redirect to home
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Check expired
	if link.ExpiredAt != nil {
		if time.Now().After(*link.ExpiredAt) {
			// Stop save this page to cache
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
			c.HTML(http.StatusGone, "expired.html", nil)
			return
		}
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

	c.Redirect(http.StatusFound, link.OriginalURL)
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

// Delete expired link in db
func CleanExpiredLinks() {
	store.DB.Where("expired_at IS NOT NULL AND expired_at < ?", time.Now()).Delete(&models.Link{})
}

// Edit link and expired
func UpdateLink(c *gin.Context) {
	id := c.Param("id")

	// Struct input
	var input struct {
		OriginalURL   string `json:"original_url"`
		DurationType  string `json:"duration_type"` // "minutes", "hours", "days", "infinite", "expired"
		DurationValue int    `json:"duration_value"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// check if link exist
	var link models.Link
	if err := store.DB.First(&link, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy liên kết"})
		return
	}

	// Handle expired
	var newExpiredAt *time.Time
	now := time.Now()

	// if duration type is empty keep the same
	if input.DurationType == "" {
		newExpiredAt = link.ExpiredAt
	} else {
		switch input.DurationType {
		case "minutes":
			t := now.Add(time.Duration(input.DurationValue) * time.Minute)
			newExpiredAt = &t
		case "hours":
			t := now.Add(time.Duration(input.DurationValue) * time.Hour)
			newExpiredAt = &t
		case "days":
			t := now.AddDate(0, 0, input.DurationValue)
			newExpiredAt = &t
		case "infinite":
			newExpiredAt = nil
		case "expired":
			t := now.Add(-1 * time.Second) // set to expired
			newExpiredAt = &t
		default:
			newExpiredAt = link.ExpiredAt
		}
	}

	// Update to db
	updates := map[string]interface{}{
		"original_url": input.OriginalURL,
		"expired_at":   newExpiredAt,
	}

	if err := store.DB.Model(&link).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể cập nhật liên kết"})
		return
	}

	// notify
	go NotifyDataChange()

	c.JSON(http.StatusOK, gin.H{
		"message":        "Cập nhật thành công",
		"new_expiration": newExpiredAt,
	})
}
