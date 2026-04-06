package handler

import (
	"net/http"

	"submanager/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createGroupRequest struct {
	Name            string `json:"name" binding:"required,max=64"`
	Description     string `json:"description" binding:"max=256"`
	SortOrder       int    `json:"sort_order"`
	Enabled         *bool  `json:"enabled"`
	SubscriptionIDs []uint `json:"subscription_ids"`
	AgentIDs        []uint `json:"agent_ids"`
}

type updateGroupRequest struct {
	Name            string `json:"name" binding:"omitempty,max=64"`
	Description     string `json:"description" binding:"omitempty,max=256"`
	SortOrder       *int   `json:"sort_order"`
	Enabled         *bool  `json:"enabled"`
	SubscriptionIDs []uint `json:"subscription_ids"`
	AgentIDs        []uint `json:"agent_ids"`
}

// ListGroups returns all service groups.
func ListGroups(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var groups []model.ServiceGroup
	db.Order("sort_order ASC, id DESC").Find(&groups)

	type groupWithRelations struct {
		model.ServiceGroup
		SubscriptionIDs []uint `json:"subscription_ids"`
		AgentIDs        []uint `json:"agent_ids"`
	}

	result := make([]groupWithRelations, 0, len(groups))
	for _, g := range groups {
		var subIDs []uint
		db.Model(&model.GroupSubscriptionSource{}).Where("service_group_id = ?", g.ID).Pluck("subscription_source_id", &subIDs)

		var agentIDs []uint
		db.Model(&model.GroupAgent{}).Where("service_group_id = ?", g.ID).Pluck("agent_id", &agentIDs)

		result = append(result, groupWithRelations{
			ServiceGroup:    g,
			SubscriptionIDs: subIDs,
			AgentIDs:        agentIDs,
		})
	}

	c.JSON(http.StatusOK, result)
}

// CreateGroup creates a new service group.
func CreateGroup(c *gin.Context) {
	var req createGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	group := model.ServiceGroup{
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		Enabled:     enabled,
	}

	db := c.MustGet("db").(*gorm.DB)
	if err := db.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create service group"})
		return
	}

	// Create relations
	for _, subID := range req.SubscriptionIDs {
		db.Create(&model.GroupSubscriptionSource{
			ServiceGroupID:       group.ID,
			SubscriptionSourceID: subID,
		})
	}
	for _, agentID := range req.AgentIDs {
		db.Create(&model.GroupAgent{
			ServiceGroupID: group.ID,
			AgentID:        agentID,
		})
	}

	c.JSON(http.StatusCreated, group)
}

// UpdateGroup updates a service group by ID.
func UpdateGroup(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var group model.ServiceGroup
	if err := db.First(&group, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "service group not found"})
		return
	}

	var req updateGroupRequest
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
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		db.Model(&group).Updates(updates)
	}

	// Update relations if provided
	if req.SubscriptionIDs != nil {
		db.Where("service_group_id = ?", group.ID).Delete(&model.GroupSubscriptionSource{})
		for _, subID := range req.SubscriptionIDs {
			db.Create(&model.GroupSubscriptionSource{
				ServiceGroupID:       group.ID,
				SubscriptionSourceID: subID,
			})
		}
	}

	if req.AgentIDs != nil {
		db.Where("service_group_id = ?", group.ID).Delete(&model.GroupAgent{})
		for _, agentID := range req.AgentIDs {
			db.Create(&model.GroupAgent{
				ServiceGroupID: group.ID,
				AgentID:        agentID,
			})
		}
	}

	db.First(&group, group.ID)
	c.JSON(http.StatusOK, group)
}

// DeleteGroup deletes a service group by ID.
func DeleteGroup(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	result := db.Delete(&model.ServiceGroup{}, id)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "service group not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "service group deleted"})
}
