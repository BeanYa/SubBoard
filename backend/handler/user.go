package handler

import (
	"net/http"

	"submanager/config"
	"submanager/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type passwordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ProfileHandler returns the current user's profile.
func ProfileHandler(c *gin.Context) {
	user := c.MustGet("user").(*model.User)

	db := c.MustGet("db").(*gorm.DB)
	result := gin.H{
		"id":               user.ID,
		"username":         user.Username,
		"is_admin":         user.IsAdmin,
		"enabled":          user.Enabled,
		"sub_token":        user.SubToken,
		"traffic_used":     user.TrafficUsed,
		"expire_at":        user.ExpireAt,
		"traffic_reset_at": user.TrafficResetAt,
		"created_at":       user.CreatedAt,
	}

	if user.PlanID != nil {
		var plan model.Plan
		if err := db.First(&plan, *user.PlanID).Error; err == nil {
			result["plan"] = gin.H{
				"id":            plan.ID,
				"name":          plan.Name,
				"traffic_limit": plan.TrafficLimit,
				"duration_days": plan.DurationDays,
			}
		}
	}

	c.JSON(http.StatusOK, result)
}

// SubscriptionHandler returns the user's subscription details.
func SubscriptionHandler(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	cfg := c.MustGet("config").(*config.Config)
	db := c.MustGet("db").(*gorm.DB)

	var trafficLimit int64
	var planName string
	if user.PlanID != nil {
		var plan model.Plan
		if err := db.First(&plan, *user.PlanID).Error; err == nil {
			trafficLimit = plan.TrafficLimit
			planName = plan.Name
		}
	}

	subURL := cfg.SubBaseURL + "/sub/" + user.SubToken

	c.JSON(http.StatusOK, gin.H{
		"plan_name":       planName,
		"traffic_used":    user.TrafficUsed,
		"traffic_limit":   trafficLimit,
		"expire_at":       user.ExpireAt,
		"sub_url":         subURL,
		"sub_url_clash":   subURL + "?format=clash",
		"sub_url_singbox": subURL + "?format=singbox",
		"sub_url_base64":  subURL + "?format=base64",
	})
}

// PasswordHandler updates the current user's password.
func PasswordHandler(c *gin.Context) {
	var req passwordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := c.MustGet("user").(*model.User)
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect current password"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	db.Model(user).Update("password_hash", string(hash))

	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}
