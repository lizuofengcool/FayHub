package model

type TenantQuota struct {
	BaseModel
	TenantID        uint  `gorm:"uniqueIndex;not null" json:"tenant_id"`
	MaxUsers        int   `gorm:"default:10;not null" json:"max_users"`
	MaxStorageMB    int   `gorm:"default:1024;not null" json:"max_storage_mb"`
	MaxPlugins      int   `gorm:"default:5;not null" json:"max_plugins"`
	MaxAPIPerDay    int   `gorm:"default:10000;not null" json:"max_api_per_day"`
	UsedUsers       int   `gorm:"default:0;not null" json:"used_users"`
	UsedStorageMB   int   `gorm:"default:0;not null" json:"used_storage_mb"`
	UsedPlugins     int   `gorm:"default:0;not null" json:"used_plugins"`
	UsedAPIPerDay   int   `gorm:"default:0;not null" json:"used_api_per_day"`
	APIResetDate    string `gorm:"size:10;default:''" json:"api_reset_date"`
}

func (TenantQuota) TableName() string {
	return "tenant_quotas"
}
