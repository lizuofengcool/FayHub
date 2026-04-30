package model

import "time"

type TokenBlacklistEntry struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	TokenHash string    `gorm:"uniqueIndex;size:64;not null" json:"token_hash"`
	ExpiresAt time.Time `gorm:"index;not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (TokenBlacklistEntry) TableName() string {
	return "token_blacklist_entries"
}
