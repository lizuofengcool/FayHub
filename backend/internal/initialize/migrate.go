package initialize

import (
	"fayhub/internal/model"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func MigrateDatabase(db *gorm.DB) error {
	log.Println("📊 开始数据库迁移...")

	models := []interface{}{
		&model.Tenant{},
		&model.User{},
		&model.TenantUser{},
		&model.Role{},
		&model.Menu{},
		&model.API{},
		&model.RoleMenu{},
		&model.RoleAPI{},
		&model.UserRole{},
		&model.TenantRole{},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", m, err)
		}
		log.Printf("✅ 迁移完成: %T", m)
	}

	log.Println("🎯 数据库迁移完成！")
	return nil
}

func InitTestData(db *gorm.DB) error {
	log.Println("📝 开始初始化测试数据...")

	var tenantCount int64
	db.Model(&model.Tenant{}).Count(&tenantCount)
	if tenantCount > 0 {
		log.Println("ℹ️  测试数据已存在，跳过初始化")
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	defaultTenant := &model.Tenant{
		Name:        "测试租户",
		Domain:      "test.fayhub.com",
		Description: "系统默认创建的测试租户",
		Status:      1,
	}
	if err := db.Create(defaultTenant).Error; err != nil {
		return fmt.Errorf("failed to create default tenant: %w", err)
	}
	log.Println("✅ 默认租户创建完成")

	adminUser := &model.User{
		Username: "admin",
		Password: string(hashedPassword),
		Email:    "admin@fayhub.com",
		Status:   1,
		Role:     "super_admin",
		RealName: "超级管理员",
	}
	if err := db.Create(adminUser).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}
	log.Println("✅ 超级管理员创建完成")

	testUser := &model.User{
		Username: "test_user",
		Password: string(hashedPassword),
		Email:    "test@fayhub.com",
		Phone:    "13800138001",
		Status:   1,
		Role:     "tenant_admin",
		RealName: "测试用户",
	}
	if err := db.Create(testUser).Error; err != nil {
		return fmt.Errorf("failed to create test user: %w", err)
	}
	log.Println("✅ 测试用户创建完成")

	log.Println("🎯 测试数据初始化完成！")
	return nil
}
