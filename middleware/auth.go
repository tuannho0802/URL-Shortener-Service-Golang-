package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("your_secret_key_123") // secret key

type MyClaims struct {
	UserID uint
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
		return &MyClaims{UserID: uID}, nil
	}
	return nil, fmt.Errorf("invalid claims")
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Lấy Token từ header "Authorization"
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort() // Stop execution
			return
		}

		// Header: "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Get user_id from token
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userID := uint(claims["user_id"].(float64))
			c.Set("userID", userID) // Save user_id to context
		}

		c.Next() // Move to the next handler
	}
}
