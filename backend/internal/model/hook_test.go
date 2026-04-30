package model

import (
	"context"
	"testing"

	"fayhub/pkg/utils"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestCallbackTenantIsolation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	err = db.AutoMigrate(&User{}, &Role{})
	if err != nil {
		t.Fatal(err)
	}

	err = RegisterTenantIsolationCallbacks(db)
	if err != nil {
		t.Fatal(err)
	}

	adminCtx := utils.SkipTenantIsolation(context.Background())
	ctxA := utils.WithTenantID(context.Background(), 1)
	ctxB := utils.WithTenantID(context.Background(), 2)

	userA := User{TenantModel: TenantModel{TenantID: 1}, Username: "test_user", Password: "hash", Status: 1}
	if err := db.WithContext(adminCtx).Create(&userA).Error; err != nil {
		t.Fatalf("Failed to create userA: %v", err)
	}

	var found User
	err = db.WithContext(ctxA).Where("username = ?", "test_user").First(&found).Error
	if err != nil {
		t.Errorf("ctxA should find userA, got error: %v", err)
	}

	var notFound User
	err = db.WithContext(ctxB).Where("username = ?", "test_user").First(&notFound).Error
	if err == nil {
		t.Errorf("ctxB should NOT find userA, but found: %+v", notFound)
	}

	roleA := Role{TenantModel: TenantModel{TenantID: 1}, Name: "test_role", Type: 2, Status: 1}
	if err := db.WithContext(adminCtx).Create(&roleA).Error; err != nil {
		t.Fatalf("Failed to create roleA: %v", err)
	}

	var foundRole Role
	err = db.WithContext(ctxA).Where("name = ?", "test_role").First(&foundRole).Error
	if err != nil {
		t.Errorf("ctxA should find roleA, got error: %v", err)
	}

	var notFoundRole Role
	err = db.WithContext(ctxB).Where("name = ?", "test_role").First(&notFoundRole).Error
	if err == nil {
		t.Errorf("ctxB should NOT find roleA, but found: %+v", notFoundRole)
	}

	var allUsers []User
	err = db.WithContext(adminCtx).Find(&allUsers).Error
	if err != nil {
		t.Errorf("adminCtx should find all users, got error: %v", err)
	}
	if len(allUsers) < 1 {
		t.Errorf("adminCtx should find at least 1 user, found %d", len(allUsers))
	}
}
