package middleware

import (
	"net/http"
	"strings"

	"submanager/config"
	"submanager/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AgentTokenAuth validates agent token from Authorization header.
// Used for agent→panel communication (report, register, etc.)
func AgentTokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := c.MustGet("db").(*gorm.DB)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header", "code": "AUTH_MISSING"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format, expected Bearer token", "code": "AUTH_INVALID_FORMAT"})
			c.Abort()
			return
		}

		token := parts[1]
		if len(token) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "empty agent token", "code": "AUTH_EMPTY_TOKEN"})
			c.Abort()
			return
		}

		var agent model.Agent
		if err := db.Where("token = ?", token).First(&agent).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid agent token", "code": "AUTH_INVALID_TOKEN"})
			c.Abort()
			return
		}

		if !agent.Enabled {
			c.JSON(http.StatusForbidden, gin.H{"error": "agent disabled", "code": "AGENT_DISABLED"})
			c.Abort()
			return
		}

		c.Set("agent", &agent)
		c.Next()
	}
}

// AgentOptionalTokenAuth is like AgentTokenAuth but does not abort on missing auth.
// Used for endpoints that accept both authenticated and unauthenticated requests.
func AgentOptionalTokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := c.MustGet("config").(*config.Config)
		_ = cfg // reserved for future use

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.Next()
			return
		}

		token := parts[1]
		db := c.MustGet("db").(*gorm.DB)

		var agent model.Agent
		if err := db.Where("token = ?", token).First(&agent).Error; err != nil {
			c.Next()
			return
		}

		if agent.Enabled {
			c.Set("agent", &agent)
		}

		c.Next()
	}
}
