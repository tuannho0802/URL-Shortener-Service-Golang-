package handlers

import (
	"net/http"
	"time"

	"github.com/tuannho0802/URL-Shortener-Service-Golang-/middleware"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/models"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("your_secret_key_123")

func Register(c *gin.Context) {
	var input struct {
		Username       string `json:"username" binding:"required"`
		Password       string `json:"password" binding:"required"`
		RetypePassword string `json:"retype_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng nhập đầy đủ thông tin"})
		return
	}

	// Check retype password
	if input.Password != input.RetypePassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mật khẩu nhập lại không khớp"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi xử lý mật khẩu"})
		return
	}

	// Create user
	user := models.User{
		Username: input.Username,
		Password: string(hashedPassword),
		Role:     "user",
	}

	if err := store.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tên đăng nhập đã tồn tại"})
		return
	}

	// notify
	NotifyDataChange(0)

	c.JSON(http.StatusOK, gin.H{"message": "Đăng ký thành công"})
}

func Login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thông tin không hợp lệ"})
		return
	}

	var user models.User
	if err := store.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Người dùng không tồn tại"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sai mật khẩu"})
		return
	}

	// Add role to token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := middleware.MyClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo mã xác thực"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng nhập thành công",
		"token":   tokenString,
		"role":    user.Role, // Add role
	})
}

// ForgotPassword (send to email later)
func ForgotPassword(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
	}
	c.ShouldBindJSON(&input)

	var user models.User
	if err := store.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "User không tồn tại"})
		return
	}

	// reset token
	resetToken := "MÃ_NGẪU_NHIÊN_123"
	user.ResetToken = resetToken
	user.ResetTokenExpiry = time.Now().Add(15 * time.Minute)
	store.DB.Save(&user)

	c.JSON(200, gin.H{"message": "Dùng mã này để reset (Sau này sẽ gửi qua email)", "token": resetToken})
}
