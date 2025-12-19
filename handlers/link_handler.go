package handlers

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/models"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/store"
)

// Hàm tạo chuỗi ngẫu nhiên cho ShortCode [cite: 24]
func generateShortCode(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// 1. API Tạo link rút gọn
func CreateShortLink(c *gin.Context) {
	var input struct {
		OriginalURL string `json:"original_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Tạo mã ngẫu nhiên 6 ký tự
	shortCode := generateShortCode(6)
	// Lưu ý: Thực tế cần check xem mã này đã tồn tại chưa để tránh trùng lặp [cite: 25]

	link := models.Link{
		OriginalURL: input.OriginalURL,
		ShortCode:   shortCode,
		CreatedAt:   time.Now(),
	}

	store.DB.Create(&link)

	// Trả về full URL rút gọn
	c.JSON(http.StatusOK, gin.H{
		"short_url": "http://localhost:8080/" + shortCode,
		"data":      link,
	})
}

// 2. API Redirect
func RedirectLink(c *gin.Context) {
	code := c.Param("code")
	var link models.Link

	// Tìm link trong DB theo ShortCode
	if err := store.DB.Where("short_code = ?", code).First(&link).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	// Tăng lượt click [cite: 14]
	link.ClickCount++
	store.DB.Save(&link)

	// Redirect về trang gốc
	c.Redirect(http.StatusFound, link.OriginalURL)
}

// 3. API Lấy danh sách link [cite: 19]
func GetAllLinks(c *gin.Context) {
	var links []models.Link
	store.DB.Find(&links)
	c.JSON(http.StatusOK, links)
}
