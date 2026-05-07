package model

import "time"

type PluginResourceStats struct {
	TenantModel
	PluginID       string    `json:"plugin_id" gorm:"size:100;index;not null"`
	CallCount      int64     `json:"call_count" gorm:"default:0"`
	ErrorCount     int64     `json:"error_count" gorm:"default:0"`
	TotalDurationMs int64    `json:"total_duration_ms" gorm:"default:0"`
	MaxDurationMs  int64     `json:"max_duration_ms" gorm:"default:0"`
	LastCallAt     time.Time `json:"last_call_at"`
	LastErrorAt    *time.Time `json:"last_error_at,omitempty"`
	LastErrorMsg   string    `json:"last_error_msg" gorm:"size:500"`
	MemoryUsageKB  int64     `json:"memory_usage_kb" gorm:"default:0"`
	Status         string    `json:"status" gorm:"size:20;default:active"`
}

func (PluginResourceStats) TableName() string {
	return "plugin_resource_stats"
}
