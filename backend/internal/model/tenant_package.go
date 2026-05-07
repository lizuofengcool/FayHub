package model

type TenantPackage struct {
	SnowflakeModel
	Name        string `json:"name" gorm:"size:100;not null"`
	Code        string `json:"code" gorm:"size:100;not null;uniqueIndex"`
	Status      int    `json:"status" gorm:"default:1"`
	Sort        int    `json:"sort" gorm:"default:0"`
	Remark      string `json:"remark" gorm:"size:500"`
	MaxUsers    int    `json:"max_users" gorm:"default:10"`
	MaxStorageMB int   `json:"max_storage_mb" gorm:"default:1024"`
	MaxPlugins  int    `json:"max_plugins" gorm:"default:5"`
}

func (TenantPackage) TableName() string {
	return "tenant_packages"
}

type TenantPackageMenu struct {
	SnowflakeModel
	PackageID int64 `json:"package_id" gorm:"index;not null"`
	MenuID    int64 `json:"menu_id" gorm:"index;not null"`
}

func (TenantPackageMenu) TableName() string {
	return "tenant_package_menus"
}
