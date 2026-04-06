package handler

import (
	"net/http"

	"submanager/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createSubscriptionRequest struct {
	Name            string        `json:"name" binding:"required,max=64"`
	Type            string        `json:"type" binding:"required,oneof=substore url raw"`
	URL             string        `json:"url"`
	Headers         model.JSONMap `json:"headers"`
	RefreshInterval int           `json:"refresh_interval"`
	Enabled         *bool         `json:"enabled"`
}

type updateSubscriptionRequest struct {
	Name            string        `json:"name" binding:"omitempty,max=64"`
	Type            string        `json:"type" binding:"omitempty,oneof=substore url raw"`
	URL             string        `json:"url"`
	Headers         model.JSONMap `json:"headers"`
	RefreshInterval *int          `json:"refresh_interval"`
	Enabled         *bool         `json:"enabled"`
}

// ListSubscriptions returns all subscription sources.
func ListSubscriptions(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var sources []model.SubscriptionSource
	db.Order("id DESC").Find(&sources)
	c.JSON(http.StatusOK, sources)
}

// CreateSubscription creates a new subscription source.
func CreateSubscription(c *gin.Context) {
	var req createSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	source := model.SubscriptionSource{
		Name:            req.Name,
		Type:            req.Type,
		URL:             req.URL,
		Headers:         req.Headers,
		RefreshInterval: req.RefreshInterval,
		Enabled:         enabled,
	}

	db := c.MustGet("db").(*gorm.DB)
	if err := db.Create(&source).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription source"})
		return
	}

	c.JSON(http.StatusCreated, source)
}

// UpdateSubscription updates a subscription source by ID.
func UpdateSubscription(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var source model.SubscriptionSource
	if err := db.First(&source, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription source not found"})
		return
	}

	var req updateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.URL != "" {
		updates["url"] = req.URL
	}
	if req.Headers != nil {
		updates["headers"] = req.Headers
	}
	if req.RefreshInterval != nil {
		updates["refresh_interval"] = *req.RefreshInterval
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		db.Model(&source).Updates(updates)
	}

	db.First(&source, source.ID)
	c.JSON(http.StatusOK, source)
}

// DeleteSubscription deletes a subscription source by ID.
func DeleteSubscription(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	result := db.Delete(&model.SubscriptionSource{}, id)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription source not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription source deleted"})
}

// RefreshSubscription manually refreshes nodes from a subscription source.
func RefreshSubscription(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var source model.SubscriptionSource
	if err := db.First(&source, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription source not found"})
		return
	}

	// TODO: trigger async subscription fetch in Phase 1 Step 3
	c.JSON(http.StatusOK, gin.H{"message": "refresh queued", "source_id": source.ID})
}
