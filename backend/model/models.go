package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// JSONMap is a helper type for storing map[string]any as JSON in GORM.
type JSONMap map[string]any

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	b, err := json.Marshal(j)
	return string(b), err
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = JSONMap{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// User 用户表
type User struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Username       string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	PasswordHash   string         `gorm:"type:varchar(256);not null" json:"-"`
	UUID           string         `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"` // sing-box multi-user identity
	PlanID         *uint          `gorm:"index" json:"plan_id"`
	Plan           Plan           `gorm:"foreignKey:PlanID;constraint:OnDelete:SET NULL" json:"plan,omitempty"`
	SubToken       string         `gorm:"type:varchar(32);uniqueIndex;not null" json:"sub_token"`
	TrafficUsed    int64          `gorm:"default:0" json:"traffic_used"`
	ExpireAt       *time.Time     `json:"expire_at"`
	TrafficResetAt *time.Time     `json:"traffic_reset_at"`
	IsAdmin        bool           `gorm:"default:false" json:"is_admin"`
	Enabled        bool           `gorm:"default:true" json:"enabled"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserTrafficLog per-user traffic audit log
type UserTrafficLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	AgentID   uint      `gorm:"index;not null" json:"agent_id"`
	Upload    int64     `gorm:"default:0" json:"upload"`
	Download  int64     `gorm:"default:0" json:"download"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

// Plan 套餐表
type Plan struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"type:varchar(64);not null" json:"name"`
	Description  string         `gorm:"type:varchar(256)" json:"description"`
	TrafficLimit int64          `gorm:"default:0" json:"traffic_limit"`
	DurationDays int            `gorm:"default:0" json:"duration_days"`
	Price        string         `json:"price"`
	Enabled      bool           `gorm:"default:true" json:"enabled"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// SubscriptionSource 订阅源表
type SubscriptionSource struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"type:varchar(64);not null" json:"name"`
	Type            string         `gorm:"type:varchar(16);not null" json:"type"` // substore / url / raw
	URL             string         `gorm:"type:varchar(512)" json:"url"`
	Headers         JSONMap        `gorm:"type:json" json:"headers"`
	RefreshInterval int            `gorm:"default:0" json:"refresh_interval"`
	NodeCount       int            `gorm:"default:0" json:"node_count"`
	LastFetchAt     *time.Time     `json:"last_fetch_at"`
	FetchError      string         `json:"fetch_error"`
	Enabled         bool           `gorm:"default:true" json:"enabled"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// ServiceGroup 服务群表
type ServiceGroup struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(64);not null" json:"name"`
	Description string         `gorm:"type:varchar(256)" json:"description"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// PlanServiceGroup 套餐 <-> 服务群关联表
type PlanServiceGroup struct {
	PlanID         uint `gorm:"primaryKey;foreignKey:PlanID;constraint:OnDelete:CASCADE" json:"plan_id"`
	ServiceGroupID uint `gorm:"primaryKey;foreignKey:ServiceGroupID;constraint:OnDelete:CASCADE" json:"service_group_id"`
}

// GroupSubscriptionSource 服务群 <-> 订阅源关联表
type GroupSubscriptionSource struct {
	ServiceGroupID       uint `gorm:"primaryKey;foreignKey:ServiceGroupID;constraint:OnDelete:CASCADE" json:"service_group_id"`
	SubscriptionSourceID uint `gorm:"primaryKey;foreignKey:SubscriptionSourceID;constraint:OnDelete:CASCADE" json:"subscription_source_id"`
}

// GroupAgent 服务群 <-> Agent 关联表
type GroupAgent struct {
	ServiceGroupID uint `gorm:"primaryKey;foreignKey:ServiceGroupID;constraint:OnDelete:CASCADE" json:"service_group_id"`
	AgentID        uint `gorm:"primaryKey;foreignKey:AgentID;constraint:OnDelete:CASCADE" json:"agent_id"`
}

// Agent VPS Agent 表
type Agent struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Name           string         `gorm:"type:varchar(64);not null" json:"name"`
	Token          string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"token"`
	ServerAddr     string         `gorm:"type:varchar(128)" json:"server_addr"`
	Port           int            `json:"port"`
	Protocol       string         `gorm:"type:varchar(16)" json:"protocol"` // vless/vmess/shadowsocks/trojan/snell/hysteria2
	ProtocolConfig JSONMap        `gorm:"type:json" json:"protocol_config"`
	TrafficUsed    int64          `gorm:"default:0" json:"traffic_used"`
	TrafficTotal   int64          `gorm:"default:0" json:"traffic_total"`
	CPUUsage       float64        `gorm:"default:0" json:"cpu_usage"`
	MemUsage       float64        `gorm:"default:0" json:"mem_usage"`
	DiskUsage      float64        `gorm:"default:0" json:"disk_usage"`
	Load1m         float64        `gorm:"default:0" json:"load_1m"`
	Status         string         `gorm:"type:varchar(16);default:'unknown'" json:"status"` // online/offline/unknown/error
	LastReportAt   *time.Time     `json:"last_report_at"`
	OnlineDevices  int            `gorm:"default:0" json:"online_devices"`
	Uptime         int64          `gorm:"default:0" json:"uptime"`
	Hostname       string         `gorm:"type:varchar(128)" json:"hostname"`
	OS             string         `gorm:"type:varchar(32)" json:"os"`
	Arch           string         `gorm:"type:varchar(32)" json:"arch"`
	IP             string         `gorm:"type:varchar(64)" json:"ip"`
	KernelVersion  string         `gorm:"type:varchar(64)" json:"kernel_version"`
	AgentVersion   string         `gorm:"type:varchar(32)" json:"agent_version"`
	CPUCores       int            `gorm:"default:0" json:"cpu_cores"`
	MemoryTotal    int64          `gorm:"default:0" json:"memory_total"`
	DiskTotal      int64          `gorm:"default:0" json:"disk_total"`
	RegisteredAt   *time.Time     `json:"registered_at"`
	ErrorMessage   string         `gorm:"type:varchar(512)" json:"error_message"`
	Enabled        bool           `gorm:"default:true" json:"enabled"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// NodeCache 节点缓存表
type NodeCache struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SourceType string    `gorm:"type:varchar(16);not null;index:idx_source" json:"source_type"` // subscription / agent
	SourceID   uint      `gorm:"not null;index:idx_source" json:"source_id"`
	Name       string    `gorm:"type:varchar(128);not null" json:"name"`
	RawLink    string    `gorm:"type:varchar(1024)" json:"raw_link"`
	Protocol   string    `gorm:"type:varchar(16)" json:"protocol"`
	Address    string    `gorm:"type:varchar(128)" json:"address"`
	Port       int       `json:"port"`
	Extra      JSONMap   `gorm:"type:json" json:"extra"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AllModels returns all models for auto-migration.
func AllModels() []interface{} {
	return []interface{}{
		&User{},
		&UserTrafficLog{},
		&Plan{},
		&SubscriptionSource{},
		&ServiceGroup{},
		&PlanServiceGroup{},
		&GroupSubscriptionSource{},
		&GroupAgent{},
		&Agent{},
		&NodeCache{},
	}
}
