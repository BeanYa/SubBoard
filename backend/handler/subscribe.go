package handler

import (
	"net/http"
	"strings"
	"time"

	"submanager/model"
	"submanager/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SubscribeHandler serves subscription output for users based on their plan.
// Route: GET /sub/:token (protected by TokenAuth middleware).
func SubscribeHandler(c *gin.Context) {
	user := c.MustGet("sub_user").(*model.User)
	db := c.MustGet("db").(*gorm.DB)

	// Check if user is enabled
	if !user.Enabled {
		c.String(http.StatusForbidden, "# Account disabled\n")
		return
	}

	// Check expiry
	if user.ExpireAt != nil && time.Now().After(*user.ExpireAt) {
		c.String(http.StatusOK, "# Subscription expired\n")
		return
	}

	// Check user has a plan
	if user.PlanID == nil {
		c.String(http.StatusOK, "# No active plan\n")
		return
	}

	// Load plan
	var plan model.Plan
	if err := db.First(&plan, *user.PlanID).Error; err != nil {
		c.String(http.StatusOK, "# Plan not found\n")
		return
	}

	// Check traffic limit (0 = unlimited)
	if plan.TrafficLimit > 0 && user.TrafficUsed >= plan.TrafficLimit {
		c.String(http.StatusOK, "# Traffic limit exceeded\n")
		return
	}

	// Collect nodes for this user
	nodes, err := service.CollectNodesForUser(db, user.ID)
	if err != nil || len(nodes) == 0 {
		c.String(http.StatusOK, "# No nodes available\n")
		return
	}

	// Detect output format
	ua := c.GetHeader("User-Agent")
	format := detectFormat(ua, c.DefaultQuery("format", ""))

	// Convert and respond
	switch format {
	case "clash":
		result := service.ConvertToClashYAML(nodes, "SubManager")
		c.Data(http.StatusOK, "text/yaml; charset=utf-8", []byte(result))
	case "singbox":
		result := service.ConvertToSingboxJSON(nodes, "SubManager")
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(result))
	case "raw":
		result := service.ConvertToRaw(nodes)
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(result))
	default: // base64
		result := service.ConvertToBase64(nodes)
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(result))
	}
}

// detectFormat determines the subscription output format from query param or User-Agent.
func detectFormat(ua, queryFormat string) string {
	if queryFormat != "" {
		f := strings.ToLower(queryFormat)
		switch f {
		case "clash", "yaml":
			return "clash"
		case "singbox", "sing-box", "json":
			return "singbox"
		case "raw":
			return "raw"
		case "base64", "b64":
			return "base64"
		}
		return f
	}

	// Auto-detect from User-Agent
	uaLower := strings.ToLower(ua)
	if strings.Contains(uaLower, "clash") {
		return "clash"
	}
	if strings.Contains(uaLower, "sing-box") || strings.Contains(uaLower, "singbox") ||
		strings.Contains(uaLower, "sfi") || strings.Contains(uaLower, "sfa") {
		return "singbox"
	}

	return "base64"
}
