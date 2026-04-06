package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"submanager/config"
	"submanager/model"
	"submanager/router"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	if err := db.AutoMigrate(model.AllModels()...); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Auto-create default admin if no admin exists
	if err := ensureDefaultAdmin(db, cfg); err != nil {
		log.Fatalf("Failed to ensure default admin: %v", err)
	}

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.Default()

	router.Setup(engine, db, cfg)

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("SubManager starting on %s (env=%s, db=%s)", addr, cfg.AppEnv, cfg.DBDriver)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	switch cfg.DBDriver {
	case "postgres":
		return gorm.Open(postgres.Open(cfg.DBDSN), &gorm.Config{})
	default:
		return gorm.Open(sqlite.Open(cfg.DBDSN), &gorm.Config{})
	}
}

// ensureDefaultAdmin creates a default admin user if no admin exists
func ensureDefaultAdmin(db *gorm.DB, cfg *config.Config) error {
	var count int64
	db.Model(&model.User{}).Where("is_admin = ?", true).Count(&count)
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	token := generateToken(32)
	user := model.User{
		Username:     cfg.AdminUsername,
		PasswordHash: string(hash),
		SubToken:     token,
		UUID:         generateUUID(),
		IsAdmin:      true,
		Enabled:      true,
	}

	if err := db.Create(&user).Error; err != nil {
		return fmt.Errorf("failed to create default admin: %w", err)
	}

	log.Printf("Default admin created: %s / %s", cfg.AdminUsername, cfg.AdminPassword)
	return nil
}

func generateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateUUID() string {
	return uuid.New().String()
}
