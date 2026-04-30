package service

import (
	"fayhub/pkg/config"
	"fayhub/pkg/utils"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func openTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	utils.SetGlobalDB(db)

	config.GlobalConfig = &config.Config{
		Security: config.SecurityConfig{
			MaxLoginAttempts: 5,
			LockDurationMin:  15,
		},
	}

	return db, nil
}
