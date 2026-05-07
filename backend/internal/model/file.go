package model

type FileRecord struct {
	SnowflakeTenantModel
	FileName      string `gorm:"type:varchar(256);not null" json:"file_name"`
	OriginalName  string `gorm:"type:varchar(256);not null" json:"original_name"`
	FileKey       string `gorm:"type:varchar(512);not null;index" json:"file_key"`
	FileSize      int64  `gorm:"type:bigint;not null" json:"file_size"`
	MimeType      string `gorm:"type:varchar(128)" json:"mime_type"`
	StorageDriver string `gorm:"type:varchar(20);default:'local'" json:"storage_driver"`
	URL           string `gorm:"type:varchar(1024)" json:"url"`
	UserID        int64  `gorm:"index" json:"user_id"`
}

func (FileRecord) TableName() string { return "file_records" }
