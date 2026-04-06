package handler

import (
	"net/http"
	"time"

	"submanager/config"
	"submanager/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// --- Request/Response types for Agent API ---

type agentRegisterRequest struct {
	Hostname      string `json:"hostname"`
	OS            string `json:"os"`
	Arch          string `json:"arch"`
	IP            string `json:"ip"`
	KernelVersion string `json:"kernel_version"`
	AgentVersion  string `json:"agent_version"`
	CPUCores      int    `json:"cpu_cores"`
	MemoryTotal   int64  `json:"memory_total"`
	DiskTotal     int64  `json:"disk_total"`
}

type agentReportRequest struct {
	Hostname      string             `json:"hostname"`
	AgentVersion  string             `json:"agent_version"`
	Uptime        int64              `json:"uptime"`
	System        agentSystemInfo    `json:"system"`
	Traffic       agentTrafficInfo   `json:"traffic"`
	UserTraffic   []agentUserTraffic `json:"user_traffic"`
	OnlineDevices int                `json:"online_devices"`
	Status        string             `json:"status"`
	ErrorMessage  string             `json:"error_message"`
	Nodes         []agentReportNode  `json:"nodes"`
}

type agentSystemInfo struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemUsage    float64 `json:"mem_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	Load1m      float64 `json:"load_1m"`
	Load5m      float64 `json:"load_5m"`
	Load15m     float64 `json:"load_15m"`
	CPUCores    int     `json:"cpu_cores"`
	MemoryTotal int64   `json:"memory_total"`
	MemoryUsed  int64   `json:"memory_used"`
	DiskTotal   int64   `json:"disk_total"`
	DiskUsed    int64   `json:"disk_used"`
}

type agentTrafficInfo struct {
	Up        int64 `json:"up"`
	Down      int64 `json:"down"`
	TotalUp   int64 `json:"total_up"`
	TotalDown int64 `json:"total_down"`
}

type agentUserTraffic struct {
	UserID   uint   `json:"user_id"`
	UUID     string `json:"uuid"`
	Upload   int64  `json:"up"`
	Download int64  `json:"down"`
}

type agentReportNode struct {
	Name     string        `json:"name"`
	Protocol string        `json:"protocol"`
	Address  string        `json:"address"`
	Port     int           `json:"port"`
	Extra    model.JSONMap `json:"extra"`
	RawLink  string        `json:"raw_link"`
}

// --- Handlers ---

// AgentRegisterHandler handles POST /api/agent/register
func AgentRegisterHandler(c *gin.Context) {
	var req agentRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "code": "BAD_REQUEST"})
		return
	}

	agent := c.MustGet("agent").(*model.Agent)
	db := c.MustGet("db").(*gorm.DB)
	cfg := c.MustGet("config").(*config.Config)

	now := time.Now()
	updates := map[string]interface{}{
		"hostname":       req.Hostname,
		"os":             req.OS,
		"arch":           req.Arch,
		"ip":             req.IP,
		"kernel_version": req.KernelVersion,
		"agent_version":  req.AgentVersion,
		"cpu_cores":      req.CPUCores,
		"memory_total":   req.MemoryTotal,
		"disk_total":     req.DiskTotal,
		"registered_at":  &now,
		"status":         "online",
	}
	db.Model(&model.Agent{}).Where("id = ?", agent.ID).Updates(updates)

	// Reload agent
	db.First(&agent, agent.ID)

	c.JSON(http.StatusOK, gin.H{
		"ok":       true,
		"agent_id": agent.ID,
		"name":     agent.Name,
		"config":   buildAgentConfig(agent, cfg),
	})
}

// AgentReportHandler handles POST /api/agent/report
func AgentReportHandler(c *gin.Context) {
	var req agentReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "code": "BAD_REQUEST"})
		return
	}

	agent := c.MustGet("agent").(*model.Agent)
	db := c.MustGet("db").(*gorm.DB)
	cfg := c.MustGet("config").(*config.Config)

	now := time.Now()

	// Update agent status
	updates := map[string]interface{}{
		"hostname":       req.Hostname,
		"agent_version":  req.AgentVersion,
		"uptime":         req.Uptime,
		"cpu_usage":      req.System.CPUUsage,
		"mem_usage":      req.System.MemUsage,
		"disk_usage":     req.System.DiskUsage,
		"load_1m":        req.System.Load1m,
		"online_devices": req.OnlineDevices,
		"status":         req.Status,
		"error_message":  req.ErrorMessage,
		"last_report_at": &now,
	}
	db.Model(&model.Agent{}).Where("id = ?", agent.ID).Updates(updates)

	// Update total traffic on agent
	if req.Traffic.Up > 0 || req.Traffic.Down > 0 {
		db.Model(&model.Agent{}).Where("id = ?", agent.ID).
			Update("traffic_used", gorm.Expr("traffic_used + ?", req.Traffic.Up+req.Traffic.Down))
	}

	// Per-user traffic processing
	for _, ut := range req.UserTraffic {
		// Update user traffic_used
		db.Model(&model.User{}).Where("id = ? AND uuid = ?", ut.UserID, ut.UUID).
			Update("traffic_used", gorm.Expr("traffic_used + ?", ut.Upload+ut.Download))

		// Write traffic log
		db.Create(&model.UserTrafficLog{
			UserID:   ut.UserID,
			AgentID:  agent.ID,
			Upload:   ut.Upload,
			Download: ut.Download,
		})
	}

	// Node cache replacement: delete old, insert new
	if len(req.Nodes) > 0 {
		db.Where("source_type = ? AND source_id = ?", "agent", agent.ID).Delete(&model.NodeCache{})
		for _, n := range req.Nodes {
			db.Create(&model.NodeCache{
				SourceType: "agent",
				SourceID:   agent.ID,
				Name:       n.Name,
				RawLink:    n.RawLink,
				Protocol:   n.Protocol,
				Address:    n.Address,
				Port:       n.Port,
				Extra:      n.Extra,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":             true,
		"interval":       cfg.AgentReportInterval,
		"config_updated": false,
		"server_time":    now.Format(time.RFC3339),
	})
}

// AgentConfigHandler handles GET /api/agent/config/:token
func AgentConfigHandler(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing token"})
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	cfg := c.MustGet("config").(*config.Config)

	var agent model.Agent
	if err := db.Where("token = ?", token).First(&agent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found", "code": "NOT_FOUND"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":              true,
		"name":            agent.Name,
		"enabled":         agent.Enabled,
		"protocol":        agent.Protocol,
		"server_addr":     agent.ServerAddr,
		"port":            agent.Port,
		"protocol_config": agent.ProtocolConfig,
		"report_interval": cfg.AgentReportInterval,
		"offline_timeout": cfg.AgentOfflineTimeout,
	})
}

// AgentInstallScriptHandler handles GET /api/agent/install.sh
func AgentInstallScriptHandler(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.String(http.StatusBadRequest, "# Error: missing agent token\n")
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	cfg := c.MustGet("config").(*config.Config)

	var agent model.Agent
	if err := db.Where("token = ?", token).First(&agent).Error; err != nil {
		c.String(http.StatusNotFound, "# Error: agent not found\n")
		return
	}

	panelURL := cfg.SubBaseURL
	script := `#!/bin/bash
set -e

echo "========================================="
echo "  SubNode Agent Installer"
echo "  Agent: ` + agent.Name + `"
echo "========================================="

# Download subnode binary
ARCH=$(uname -m)
case $ARCH in
    x86_64) BINARY_ARCH="amd64" ;;
    aarch64) BINARY_ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "Downloading subnode-linux-${BINARY_ARCH}..."
wget -q "https://github.com/your-org/subnode/releases/latest/download/subnode-linux-${BINARY_ARCH}" -O /usr/local/bin/subnode
chmod +x /usr/local/bin/subnode

# Create config directory
mkdir -p /etc/subnode

# Write config
cat > /etc/subnode/config.yaml << CONFEOF
panel:
  url: "` + panelURL + `"
  token: "` + token + `"
  report_interval: 60
  sync_interval: 300
  tls_skip_verify: false

proxy:
  enabled: true
  core: "sing-box"
  binary: "/usr/local/bin/sing-box"

subscription:
  enabled: true
  listen: "0.0.0.0:2080"

log:
  level: "info"
  path: "/var/log/subnode/agent.log"
CONFEOF

echo "Configuration written to /etc/subnode/config.yaml"
echo ""
echo "To start the agent, run:"
echo "  subnode run -c /etc/subnode/config.yaml"
echo ""
echo "Or register as systemd service for auto-start."
`

	c.Header("Content-Type", "text/x-shellscript; charset=utf-8")
	c.String(http.StatusOK, script)
}

// --- Helpers ---

func buildAgentConfig(agent *model.Agent, cfg *config.Config) gin.H {
	config := gin.H{
		"protocol":        agent.Protocol,
		"server_addr":     agent.ServerAddr,
		"port":            agent.Port,
		"protocol_config": agent.ProtocolConfig,
		"report_interval": cfg.AgentReportInterval,
		"offline_timeout": cfg.AgentOfflineTimeout,
	}

	// Extract argo config from ProtocolConfig if present
	config["argo"] = extractNested(agent.ProtocolConfig, "argo")
	config["cdn"] = extractNested(agent.ProtocolConfig, "cdn")

	return config
}

func extractNested(m model.JSONMap, key string) gin.H {
	if m == nil {
		return gin.H{}
	}
	val, ok := m[key]
	if !ok {
		return gin.H{}
	}
	nested, ok := val.(map[string]interface{})
	if !ok {
		return gin.H{}
	}
	return gin.H(nested)
}
