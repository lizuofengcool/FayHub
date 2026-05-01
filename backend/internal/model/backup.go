package model

import "time"

type BackupRecord struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Filename  string    `json:"filename" gorm:"size:255;not null"`
	FileSize  int64     `json:"file_size"`
	Status    string    `json:"status" gorm:"size:20;default:pending"`
	CreatedAt time.Time `json:"created_at"`
}

func (BackupRecord) TableName() string {
	return "backup_records"
}
