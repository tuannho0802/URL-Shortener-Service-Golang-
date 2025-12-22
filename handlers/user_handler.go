package handlers

import (
	"fmt"
	"net/http"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/models"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/store"
)

type UserResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	LinkCount int64  `json:"link_count"`
}

// GetAllUsers optimization
func GetAllUsers(c *gin.Context) {
	var results []UserResponse

	// get data in 1 query
	err := store.DB.Model(&models.User{}).
		Select("users.id, users.username, users.role, count(links.id) as link_count").
		Joins("left join links on links.user_id = users.id").
		Group("users.id").
		Scan(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách người dùng"})
		return
	}

	c.JSON(http.StatusOK, results)
}

func UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// find user to update
	var userToUpdate models.User
	if err := store.DB.First(&userToUpdate, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy người dùng"})
		return
	}

	// 2. block if username is admin
	if userToUpdate.Username == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Không thể thay đổi quyền của tài khoản Admin hệ thống!"})
		return
	}

	// 3. block root admin change own role
	currentAdminID, _ := c.Get("userID")
	if id == fmt.Sprintf("%v", currentAdminID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không thể tự hạ quyền của chính mình"})
		return
	}

	// update
	if err := store.DB.Model(&userToUpdate).Update("role", input.Role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cập nhật thất bại"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật quyền thành công"})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	currentAdminID, _ := c.Get("userID")

	// block  admin delete own account
	if id == fmt.Sprintf("%v", currentAdminID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Không thể xóa chính tài khoản đang đăng nhập"})
		return
	}

	// find and check username
	var userToDelete models.User
	if err := store.DB.Unscoped().First(&userToDelete, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy người dùng"})
		return
	}

	// protect root admin
	if userToDelete.Username == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Không thể xóa tài khoản Admin gốc của hệ thống!"})
		return
	}

	// delete
	err := store.DB.Transaction(func(tx *gorm.DB) error {
		// delete all that user's link
		if err := tx.Unscoped().Where("user_id = ?", id).Delete(&models.Link{}).Error; err != nil {
			return err
		}

		// delete user
		if err := tx.Unscoped().Delete(&userToDelete).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Xóa thất bại: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa người dùng và tất cả dữ liệu liên quan thành công"})
}
