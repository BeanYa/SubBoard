package router

import (
	"submanager/config"
	"submanager/handler"
	"submanager/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(engine *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// Inject db and cfg into handlers via context
	engine.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Set("config", cfg)
		c.Next()
	})

	// Health check
	engine.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// System init
	engine.POST("/api/system/init", handler.InitHandler)

	// Auth
	auth := engine.Group("/api/auth")
	{
		auth.POST("/login", handler.LoginHandler)
		auth.POST("/register", handler.RegisterHandler)
	}

	// User (JWT protected)
	user := engine.Group("/api/user")
	user.Use(middleware.JWTAuth())
	{
		user.GET("/profile", handler.ProfileHandler)
		user.GET("/subscription", handler.SubscriptionHandler)
		user.PUT("/password", handler.PasswordHandler)
	}

	// Subscribe output (token auth)
	engine.GET("/sub/:token", middleware.TokenAuth(), handler.SubscribeHandler)

	// Admin (JWT + admin)
	admin := engine.Group("/api/admin")
	admin.Use(middleware.JWTAuth(), middleware.AdminAuth())
	{
		// Users
		admin.GET("/users", handler.ListUsers)
		admin.POST("/users", handler.CreateUser)
		admin.PUT("/users/:id", handler.UpdateUser)
		admin.DELETE("/users/:id", handler.DeleteUser)
		admin.POST("/users/:id/plan", handler.AssignPlan)
		admin.POST("/users/:id/reset", handler.ResetTraffic)
		admin.POST("/users/:id/toggle", handler.ToggleUser)

		// Plans
		admin.GET("/plans", handler.ListPlans)
		admin.POST("/plans", handler.CreatePlan)
		admin.PUT("/plans/:id", handler.UpdatePlan)
		admin.DELETE("/plans/:id", handler.DeletePlan)

		// Subscription sources
		admin.GET("/subscriptions", handler.ListSubscriptions)
		admin.POST("/subscriptions", handler.CreateSubscription)
		admin.PUT("/subscriptions/:id", handler.UpdateSubscription)
		admin.DELETE("/subscriptions/:id", handler.DeleteSubscription)
		admin.POST("/subscriptions/:id/refresh", handler.RefreshSubscription)

		// Service groups
		admin.GET("/groups", handler.ListGroups)
		admin.POST("/groups", handler.CreateGroup)
		admin.PUT("/groups/:id", handler.UpdateGroup)
		admin.DELETE("/groups/:id", handler.DeleteGroup)

		// Agents
		admin.GET("/agents", handler.ListAgents)
		admin.POST("/agents", handler.CreateAgent)
		admin.PUT("/agents/:id", handler.UpdateAgent)
		admin.DELETE("/agents/:id", handler.DeleteAgent)
		admin.POST("/agents/:id/install-command", handler.AgentInstallCommandHandler)
		admin.POST("/agents/:id/reset-token", handler.AgentResetTokenHandler)
		admin.GET("/agents/status", handler.AgentsStatusHandler)
	}

	// Agent API (AgentTokenAuth for report/register, URL token for config)
	agent := engine.Group("/api/agent")
	{
		agent.POST("/report", middleware.AgentTokenAuth(), handler.AgentReportHandler)
		agent.POST("/register", middleware.AgentTokenAuth(), handler.AgentRegisterHandler)
		agent.GET("/config/:token", handler.AgentConfigHandler)
		agent.GET("/install.sh", handler.AgentInstallScriptHandler)
		agent.GET("/users", middleware.AgentTokenAuth(), handler.AgentUsersHandler)
	}
}
