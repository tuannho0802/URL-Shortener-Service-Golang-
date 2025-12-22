package handlers

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/models"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/store"
)

// Click count info
type clickInfo struct {
	LinkID  uint
	UserID  uint
	Browser string
	OS      string
}

// Channel to store click info
var clickChannel = make(chan clickInfo, 10000)

// Run worker in goroutine
func StartClickWorker(ctx context.Context) {
	clickWorker(ctx)
}

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

	val, _ := c.Get("userID")
	userID := val.(uint)
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
		UserID:      userID,
	}

	if err := store.DB.Create(&newLink).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lưu vào database"})
		return
	}

	NotifyDataChange(userID)

	c.JSON(200, gin.H{"short_url": "http://localhost:8080/" + code})
}

// API Redirect
func RedirectLink(c *gin.Context) {
	code := c.Param("code")

	// Remove / and index.html
	if code == "" || code == "/" || code == "index.html" || code == "favicon.ico" {
		return
	}

	var link models.Link
	// Just get original_url, expired_at, id
	if err := store.DB.Select("original_url", "expired_at", "id", "user_id").Where("short_code = ?", code).First(&link).Error; err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Check expired
	if link.ExpiredAt != nil && time.Now().After(*link.ExpiredAt) {
		c.HTML(http.StatusGone, "expired.html", nil)
		return
	}

	// Handle user agent
	ua := c.GetHeader("User-Agent")
	browser, os := parseUserAgent(ua)

	// Update optimize for db
	// Automation click count and last browser
	// Run in goroutine for client not wait for db
	clickChannel <- clickInfo{
		LinkID:  link.ID,
		UserID:  link.UserID,
		Browser: browser,
		OS:      os,
	}
	// Redirect
	c.Redirect(http.StatusFound, link.OriginalURL)
}

// Func quick parse user agent
func parseUserAgent(ua string) (string, string) {
	browser := "Other"
	os := "Other"
	if strings.Contains(ua, "Firefox") {
		browser = "Firefox"
	} else if strings.Contains(ua, "Chrome") {
		browser = "Chrome"
	} else if strings.Contains(ua, "Safari") {
		browser = "Safari"
	} else if strings.Contains(ua, "Edg") {
		browser = "Edge"
	}

	if strings.Contains(ua, "Windows") {
		os = "Windows"
	} else if strings.Contains(ua, "Macintosh") {
		os = "MacOS"
	} else if strings.Contains(ua, "Android") {
		os = "Android"
	} else if strings.Contains(ua, "iPhone") {
		os = "iOS"
	}

	return browser, os
}

// API get all list
// Update: not get all for browser safety (limit and pagination)
func GetMyLinks(c *gin.Context) {

	val, _ := c.Get("userID")
	userID := val.(uint)
	var links []models.Link

	// Get page analysis parameters from URL
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	sortParam := c.DefaultQuery("sort", "created_at_desc")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Calculate total
	var total int64
	// count only user links
	store.DB.Model(&models.Link{}).Where("user_id = ?", userID).Count(&total)

	// Access pagination (user only)
	query := store.DB.Where("user_id = ?", userID).Limit(limit).Offset(offset)

	switch sortParam {
	case "abc":
		query = query.Order("short_code ASC")
	case "clicks":
		query = query.Order("click_count DESC")
	case "oldest":
		query = query.Order("created_at ASC")
	default:
		query = query.Order("created_at DESC")
	}

	if err := query.Find(&links).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy vấn"})
		return
	}

	// Return data and pagination
	c.JSON(http.StatusOK, gin.H{
		"data":      links,
		"total":     total,
		"page":      page,
		"limit":     limit,
		"last_page": int(math.Ceil(float64(total) / float64(limit))),
	})
}

// Delete link
func DeleteLink(c *gin.Context) {
	id := c.Param("id")
	val, _ := c.Get("userID")
	userID := val.(uint)

	// Remove in db with condition user_id = userID and id = id
	result := store.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Link{})

	if result.RowsAffected == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền xóa link này"})
		return
	}

	NotifyDataChange(userID)

	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa liên kết thành công"})
}

// Delete expired link in db
func CleanExpiredLinks() {
	store.DB.Where("expired_at IS NOT NULL AND expired_at < ?", time.Now()).Delete(&models.Link{})
}

// Edit link and expired
func UpdateLink(c *gin.Context) {
	id := c.Param("id")
	val, _ := c.Get("userID")
	userID := val.(uint)

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
	if err := store.DB.Where("id = ? AND user_id = ?", id, userID).First(&link).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link không tồn tại hoặc bạn không có quyền sửa"})
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

	NotifyDataChange(userID)

	c.JSON(http.StatusOK, gin.H{
		"message":        "Cập nhật thành công",
		"new_expiration": newExpiredAt,
	})

}

// Collect all click for 2s and update

func clickWorker(ctx context.Context) {
	// Create ticker 2s
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker: Nhận tín hiệu tắt, đang lưu dữ liệu cuối cùng...")
			processRemainingClicks()
			return
		case <-ticker.C:
			// 2s handle click and notify
			processRemainingClicks()
		}
	}
}

// Handle remaining clicks
func processRemainingClicks() {
	n := len(clickChannel)
	if n == 0 {
		return
	}

	batch := make(map[uint]int)
	userToNotify := make(map[uint]bool) // list of link to notify
	var lastBrowser, lastOS string

	for i := 0; i < n; i++ {
		info := <-clickChannel
		batch[info.LinkID]++
		userToNotify[info.UserID] = true // note that user to notify
		lastBrowser = info.Browser
		lastOS = info.OS
	}

	// update db
	for id, count := range batch {
		store.DB.Model(&models.Link{}).Where("id = ?", id).Updates(map[string]interface{}{
			"click_count":  store.DB.Raw("click_count + ?", count),
			"last_browser": lastBrowser,
			"last_os":      lastOS,
		})
	}

	// send notify
	for userID := range userToNotify {
		NotifyDataChange(userID)
	}
}
