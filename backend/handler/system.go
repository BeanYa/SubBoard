package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"submanager/config"
	"submanager/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type initRequest struct {
	Username  string `json:"username" binding:"required,min=2,max=64"`
	Password  string `json:"password" binding:"required,min=6"`
	InitToken string `json:"init_token" binding:"required"`
}

// InitHandler creates the initial admin account (one-time only).
func InitHandler(c *gin.Context) {
	var req initRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := c.MustGet("config").(*config.Config)

	if cfg.InitToken == "" || req.InitToken != cfg.InitToken {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid init token"})
		return
	}

	db := c.MustGet("db").(*gorm.DB)

	var count int64
	db.Model(&model.User{}).Where("is_admin = ?", true).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "admin already exists"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	token := generateToken(32)
	user := model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		SubToken:     token,
		UUID:         generateUUID(),
		IsAdmin:      true,
		Enabled:      true,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create admin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "admin initialized",
		"username": user.Username,
	})
}

func generateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// generateUUID creates a new UUID string for sing-box multi-user identity.
func generateUUID() string {
	return uuid.New().String()
}
