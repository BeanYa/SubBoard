package handler

import (
	"net/http"
	"strconv"

	"submanager/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,min=2,max=64"`
	Password string `json:"password" binding:"required,min=6"`
}

type updateUserRequest struct {
	Username string `json:"username" binding:"omitempty,min=2,max=64"`
	Password string `json:"password" binding:"omitempty,min=6"`
}

type assignPlanRequest struct {
	PlanID *uint `json:"plan_id"`
}

// ListUsers returns a paginated list of users.
func ListUsers(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := db.Model(&model.User{})
	if search != "" {
		query = query.Where("username LIKE ?", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var users []model.User
	query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users)

	c.JSON(http.StatusOK, gin.H{
		"items":     users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateUser creates a new user.
func CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := c.MustGet("db").(*gorm.DB)

	var count int64
	db.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		SubToken:     generateToken(32),
		UUID:         generateUUID(),
		Enabled:      true,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// UpdateUser updates a user by ID.
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		updates["password_hash"] = string(hash)
	}

	if len(updates) > 0 {
		db.Model(&user).Updates(updates)
	}

	db.First(&user, user.ID)
	c.JSON(http.StatusOK, user)
}

// DeleteUser deletes a user by ID.
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	result := db.Delete(&model.User{}, id)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

// AssignPlan assigns a plan to a user.
func AssignPlan(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req assignPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.PlanID != nil {
		var plan model.Plan
		if err := db.First(&plan, *req.PlanID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "plan not found"})
			return
		}
	}

	db.Model(&user).Update("plan_id", req.PlanID)
	db.First(&user, user.ID)
	c.JSON(http.StatusOK, user)
}

// ResetTraffic resets a user's traffic usage.
func ResetTraffic(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	result := db.Model(&model.User{}).Where("id = ?", id).Update("traffic_used", 0)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "traffic reset"})
}

// ToggleUser enables or disables a user.
func ToggleUser(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	db.Model(&user).Update("enabled", !user.Enabled)
	db.First(&user, user.ID)
	c.JSON(http.StatusOK, user)
}
