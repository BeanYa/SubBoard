package handler

import (
	"net/http"
	"time"

	"submanager/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// agentUserInfo represents a user entry in the agent users list response.
type agentUserInfo struct {
	UserID       uint   `json:"user_id"`
	UUID         string `json:"uuid"`
	Enabled      bool   `json:"enabled"`
	TrafficLimit int64  `json:"traffic_limit"`
	TrafficUsed  int64  `json:"traffic_used"`
	Expired      bool   `json:"expired"`
}

// AgentUsersHandler handles GET /api/agent/users
// Returns the list of users who have access to this agent's nodes.
func AgentUsersHandler(c *gin.Context) {
	agent := c.MustGet("agent").(*model.Agent)
	db := c.MustGet("db").(*gorm.DB)

	// Find service groups this agent belongs to
	var groupAgents []model.GroupAgent
	if err := db.Where("agent_id = ?", agent.ID).Find(&groupAgents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query agent groups"})
		return
	}

	var sgIDs []uint
	for _, ga := range groupAgents {
		sgIDs = append(sgIDs, ga.ServiceGroupID)
	}

	if len(sgIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"ok":         true,
			"users":      []agentUserInfo{},
			"updated_at": time.Now().Format(time.RFC3339),
		})
		return
	}

	// Find plans that include these service groups
	var psgs []model.PlanServiceGroup
	if err := db.Where("service_group_id IN ?", sgIDs).Find(&psgs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query plan groups"})
		return
	}

	var planIDs []uint
	for _, psg := range psgs {
		planIDs = append(planIDs, psg.PlanID)
	}

	if len(planIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"ok":         true,
			"users":      []agentUserInfo{},
			"updated_at": time.Now().Format(time.RFC3339),
		})
		return
	}

	// Query users with those plans
	var users []model.User
	db.Where("plan_id IN ? AND enabled = ?", planIDs, true).Find(&users)

	// Build response with traffic limit from plan
	var result []agentUserInfo
	now := time.Now()
	for _, u := range users {
		var trafficLimit int64
		var expired bool

		if u.PlanID != nil {
			var plan model.Plan
			if err := db.First(&plan, *u.PlanID).Error; err == nil {
				trafficLimit = plan.TrafficLimit
			}
		}

		// Check if user is expired
		if u.ExpireAt != nil && u.ExpireAt.Before(now) {
			expired = true
		}
		// Check if traffic exceeded
		if trafficLimit > 0 && u.TrafficUsed >= trafficLimit {
			expired = true
		}

		enabled := u.Enabled && !expired
		result = append(result, agentUserInfo{
			UserID:       u.ID,
			UUID:         u.UUID,
			Enabled:      enabled,
			TrafficLimit: trafficLimit,
			TrafficUsed:  u.TrafficUsed,
			Expired:      expired,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"users":      result,
		"updated_at": time.Now().Format(time.RFC3339),
	})
}
