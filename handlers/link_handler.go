package handlers

import (
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/models"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/store"
)

// Hàm tạo chuỗi ngẫu nhiên cho ShortCode [cite: 24]
func GenerateShortCode(n int) string {
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
        OriginalURL string `json:"original_url"`
        CustomAlias string `json:"custom_alias"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(400, gin.H{"error": "Dữ liệu không hợp lệ"})
        return
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
    }
    store.DB.Create(&newLink)

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

	// Update click
	link.ClickCount++
	store.DB.Save(&link)

	// Redirect main page
	c.Redirect(http.StatusFound, link.OriginalURL)
}

// 3. API get list
func GetAllLinks(c *gin.Context) {
	var links []models.Link
	store.DB.Find(&links)
	c.JSON(http.StatusOK, links)
}
