package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"submanager/config"
	"submanager/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createAgentRequest struct {
	Name           string        `json:"name" binding:"required,max=64"`
	ServerAddr     string        `json:"server_addr"`
	Port           int           `json:"port"`
	Protocol       string        `json:"protocol" binding:"required,oneof=vless vmess shadowsocks trojan snell hysteria2"`
	ProtocolConfig model.JSONMap `json:"protocol_config"`
	TrafficTotal   int64         `json:"traffic_total"`
	Enabled        *bool         `json:"enabled"`
}

type updateAgentRequest struct {
	Name           string        `json:"name" binding:"omitempty,max=64"`
	ServerAddr     string        `json:"server_addr"`
	Port           *int          `json:"port"`
	Protocol       string        `json:"protocol" binding:"omitempty,oneof=vless vmess shadowsocks trojan snell hysteria2"`
	ProtocolConfig model.JSONMap `json:"protocol_config"`
	TrafficTotal   *int64        `json:"traffic_total"`
	Enabled        *bool         `json:"enabled"`
}

type installCommandRequest struct {
	Arch string `json:"arch"`
}

// ListAgents returns all agents.
func ListAgents(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var agents []model.Agent
	db.Order("id DESC").Find(&agents)
	c.JSON(http.StatusOK, agents)
}

// CreateAgent creates a new agent with auto-generated token.
func CreateAgent(c *gin.Context) {
	var req createAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	agent := model.Agent{
		Name:           req.Name,
		Token:          generateAgentToken(),
		ServerAddr:     req.ServerAddr,
		Port:           req.Port,
		Protocol:       req.Protocol,
		ProtocolConfig: req.ProtocolConfig,
		TrafficTotal:   req.TrafficTotal,
		Status:         "unknown",
		Enabled:        enabled,
	}

	db := c.MustGet("db").(*gorm.DB)
	if err := db.Create(&agent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create agent"})
		return
	}

	c.JSON(http.StatusCreated, agent)
}

// UpdateAgent updates an agent by ID.
func UpdateAgent(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var agent model.Agent
	if err := db.First(&agent, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	var req updateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.ServerAddr != "" {
		updates["server_addr"] = req.ServerAddr
	}
	if req.Port != nil {
		updates["port"] = *req.Port
	}
	if req.Protocol != "" {
		updates["protocol"] = req.Protocol
	}
	if req.ProtocolConfig != nil {
		updates["protocol_config"] = req.ProtocolConfig
	}
	if req.TrafficTotal != nil {
		updates["traffic_total"] = *req.TrafficTotal
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		db.Model(&agent).Updates(updates)
	}

	db.First(&agent, agent.ID)
	c.JSON(http.StatusOK, agent)
}

// DeleteAgent deletes an agent by ID.
func DeleteAgent(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	// Delete associated node_cache entries
	db.Where("source_type = ? AND source_id = ?", "agent", id).Delete(&model.NodeCache{})
	// Delete group associations
	db.Where("agent_id = ?", id).Delete(&model.GroupAgent{})

	result := db.Delete(&model.Agent{}, id)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "agent deleted"})
}

// AgentInstallCommandHandler generates an install command for an agent.
func AgentInstallCommandHandler(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)
	cfg := c.MustGet("config").(*config.Config)

	var agent model.Agent
	if err := db.First(&agent, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	var req installCommandRequest
	c.ShouldBindJSON(&req)
	if req.Arch == "" {
		req.Arch = "linux-amd64"
	}

	panelURL := cfg.SubBaseURL
	scriptURL := panelURL + "/api/agent/install.sh?token=" + agent.Token + "&arch=" + req.Arch
	command := "bash <(curl -sL " + scriptURL + ") -t " + agent.Token + " -u " + panelURL

	manualConfig := "panel:\n  url: " + panelURL + "\n  token: " + agent.Token

	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"command":       command,
		"script_url":    scriptURL,
		"manual_config": manualConfig,
	})
}

// AgentResetTokenHandler resets an agent's token.
func AgentResetTokenHandler(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var agent model.Agent
	if err := db.First(&agent, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	newToken := generateAgentToken()
	db.Model(&agent).Update("token", newToken)
	db.First(&agent, agent.ID)

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"token":   newToken,
		"message": "Agent token reset. Old token invalidated.",
	})
}

// AgentsStatusHandler returns an overview of all agents' statuses.
func AgentsStatusHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	cfg := c.MustGet("config").(*config.Config)

	var agents []model.Agent
	db.Find(&agents)

	total := len(agents)
	online := 0
	offline := 0
	errorCount := 0

	timeout := time.Duration(cfg.AgentOfflineTimeout) * time.Second
	now := time.Now()

	type agentStatus struct {
		ID            uint    `json:"id"`
		Name          string  `json:"name"`
		Status        string  `json:"status"`
		LastReportAt  *string `json:"last_report_at"`
		CPUUsage      float64 `json:"cpu_usage"`
		MemUsage      float64 `json:"mem_usage"`
		OnlineDevices int     `json:"online_devices"`
		TrafficUsed   int64   `json:"traffic_used"`
	}

	statuses := make([]agentStatus, 0, total)
	for _, a := range agents {
		status := a.Status
		if a.LastReportAt != nil && now.Sub(*a.LastReportAt) > timeout {
			status = "offline"
		}

		switch status {
		case "online":
			online++
		case "offline":
			offline++
		case "error":
			errorCount++
		default:
			offline++
		}

		var lastReport *string
		if a.LastReportAt != nil {
			s := a.LastReportAt.Format("2006-01-02T15:04:05Z")
			lastReport = &s
		}

		statuses = append(statuses, agentStatus{
			ID:            a.ID,
			Name:          a.Name,
			Status:        status,
			LastReportAt:  lastReport,
			CPUUsage:      a.CPUUsage,
			MemUsage:      a.MemUsage,
			OnlineDevices: a.OnlineDevices,
			TrafficUsed:   a.TrafficUsed,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"total":   total,
		"online":  online,
		"offline": offline,
		"error":   errorCount,
		"agents":  statuses,
	})
}

// generateAgentToken creates a 64-char hex token for agent authentication.
func generateAgentToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
