package handler

import (
	"net/http"

	"submanager/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createPlanRequest struct {
	Name         string `json:"name" binding:"required,max=64"`
	Description  string `json:"description" binding:"max=256"`
	TrafficLimit int64  `json:"traffic_limit"`
	DurationDays int    `json:"duration_days"`
	Price        string `json:"price"`
	Enabled      *bool  `json:"enabled"`
}

type updatePlanRequest struct {
	Name         string `json:"name" binding:"omitempty,max=64"`
	Description  string `json:"description" binding:"omitempty,max=256"`
	TrafficLimit *int64 `json:"traffic_limit"`
	DurationDays *int   `json:"duration_days"`
	Price        string `json:"price"`
	Enabled      *bool  `json:"enabled"`
}

// ListPlans returns all plans.
func ListPlans(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var plans []model.Plan
	db.Order("id DESC").Find(&plans)
	c.JSON(http.StatusOK, plans)
}

// CreatePlan creates a new plan.
func CreatePlan(c *gin.Context) {
	var req createPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	plan := model.Plan{
		Name:         req.Name,
		Description:  req.Description,
		TrafficLimit: req.TrafficLimit,
		DurationDays: req.DurationDays,
		Price:        req.Price,
		Enabled:      enabled,
	}

	db := c.MustGet("db").(*gorm.DB)
	if err := db.Create(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create plan"})
		return
	}

	c.JSON(http.StatusCreated, plan)
}

// UpdatePlan updates a plan by ID.
func UpdatePlan(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var plan model.Plan
	if err := db.First(&plan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}

	var req updatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.TrafficLimit != nil {
		updates["traffic_limit"] = *req.TrafficLimit
	}
	if req.DurationDays != nil {
		updates["duration_days"] = *req.DurationDays
	}
	if req.Price != "" {
		updates["price"] = req.Price
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		db.Model(&plan).Updates(updates)
	}

	db.First(&plan, plan.ID)
	c.JSON(http.StatusOK, plan)
}

// DeletePlan deletes a plan by ID.
func DeletePlan(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	result := db.Delete(&model.Plan{}, id)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "plan deleted"})
}
