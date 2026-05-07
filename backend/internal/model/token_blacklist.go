package model

import "time"

type TokenBlacklistEntry struct {
	BaseModel
	TokenHash string    `gorm:"uniqueIndex;size:64;not null" json:"token_hash"`
	ExpiresAt time.Time `gorm:"index;not null" json:"expires_at"`
}

func (TokenBlacklistEntry) TableName() string {
	return "token_blacklist_entries"
}
