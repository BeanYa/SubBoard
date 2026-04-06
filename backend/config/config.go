package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv              string
	AppPort             string
	AppSecret           string
	InitToken           string
	DBDriver            string
	DBDSN               string
	SubBaseURL          string
	SubRefreshInterval  int
	AllowRegister       bool
	AdminUsername       string
	AdminPassword       string
	AgentReportInterval int
	AgentOfflineTimeout int
	CORSOrigins         string
}

func LoadConfig() *Config {
	return &Config{
		AppEnv:              getEnv("APP_ENV", "development"),
		AppPort:             getEnv("APP_PORT", "8080"),
		AppSecret:           getEnv("APP_SECRET", ""),
		InitToken:           getEnv("INIT_TOKEN", ""),
		DBDriver:            getEnv("DB_DRIVER", "sqlite"),
		DBDSN:               getEnv("DB_DSN", "submanager.db"),
		SubBaseURL:          getEnv("SUB_BASE_URL", "http://localhost:8080"),
		SubRefreshInterval:  getEnvInt("SUB_REFRESH_INTERVAL", 30),
		AllowRegister:       getEnvBool("ALLOW_REGISTER", true),
		AdminUsername:       getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:       getEnv("ADMIN_PASSWORD", "admin"),
		AgentReportInterval: getEnvInt("AGENT_REPORT_INTERVAL", 60),
		AgentOfflineTimeout: getEnvInt("AGENT_OFFLINE_TIMEOUT", 180),
		CORSOrigins:         getEnv("CORS_ORIGINS", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
