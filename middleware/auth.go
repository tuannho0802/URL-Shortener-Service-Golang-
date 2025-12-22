package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/models"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/store"
)

var jwtKey = []byte("your_secret_key_123") // secret key

type MyClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// validate token
func ValidateToken(tokenString string) (*MyClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// check user_id exists
		val, ok := claims["user_id"]
		if !ok {
			return nil, fmt.Errorf("user_id missing")
		}

		// safe force
		var uID uint
		switch v := val.(type) {
		case float64:
			uID = uint(v)
		case float32:
			uID = uint(v)
		case int:
			uID = uint(v)
		default:
			return nil, fmt.Errorf("invalid user_id type")
		}

		// check role exists
		role, ok := claims["role"].(string)
		if !ok {
			role = "user" // default if missing
		}

		return &MyClaims{
			UserID: uID,
			Role:   role,
		}, nil
	}
	return nil, fmt.Errorf("invalid claims")
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get token from auth
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Yêu cầu mã xác thực (Authorization header)"})
			c.Abort()
			return
		}

		// Header: "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// parse and validate

		token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		// check token
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Mã xác thực không hợp lệ hoặc đã hết hạn"})
			c.Abort()
			return
		}

		// get data from my claims
		if claims, ok := token.Claims.(*MyClaims); ok {
			// save UserID to Context
			c.Set("userID", claims.UserID)

			// save role to Context
			c.Set("userRole", claims.Role)
		}

		c.Next()
	}
}

// admin check
func AdminCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get user id from context
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin xác thực"})
			c.Abort()
			return
		}

		// fetch db to get to newest role
		var user models.User
		if err := store.DB.Select("role", "username").First(&user, userID).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Người dùng không tồn tại"})
			c.Abort()
			return
		}

		// check role
		if user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền truy cập khu vực Admin"})
			c.Abort()
			return
		}

		// save role to context
		c.Set("userRole", user.Role)
		c.Next()
	}
}
