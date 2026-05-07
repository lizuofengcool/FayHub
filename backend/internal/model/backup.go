package model

import "time"

type BackupRecord struct {
	ID        int64      `json:"id" gorm:"primaryKey"`
	Filename  string    `json:"filename" gorm:"size:255;not null"`
	FileSize  int64     `json:"file_size"`
	Status    string    `json:"status" gorm:"size:20;default:pending"`
	Notes     string    `json:"notes" gorm:"size:500"`
	Volumes   int       `json:"volumes" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
}

func (BackupRecord) TableName() string {
	return "backup_records"
}
