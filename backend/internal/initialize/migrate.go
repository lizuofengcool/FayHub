package initialize

import (
	"context"
	"fayhub/internal/model"
	"fayhub/pkg/utils"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func MigrateDatabase(db *gorm.DB) error {
	log.Println("开始数据库迁移...")

	ctx := utils.SkipTenantIsolation(context.Background())
	db = db.WithContext(ctx)

	// 先确保表有必要的字段
	log.Println("检查并添加必要的字段...")
	if err := db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 0").Error; err != nil {
		log.Printf("添加users.tenant_id字段失败(可能已存在): %v", err)
	} else {
		log.Println("✅ users.tenant_id字段检查/添加成功")
	}

	if err := db.Exec("ALTER TABLE tenant_users ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 0").Error; err != nil {
		log.Printf("添加tenant_users.tenant_id字段失败(可能已存在): %v", err)
	} else {
		log.Println("✅ tenant_users.tenant_id字段检查/添加成功")
	}

	if err := db.Exec("ALTER TABLE roles ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 0").Error; err != nil {
		log.Printf("添加roles.tenant_id字段失败(可能已存在): %v", err)
	} else {
		log.Println("✅ roles.tenant_id字段检查/添加成功")
	}

	if err := db.Exec("ALTER TABLE role_menus ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 0").Error; err != nil {
		log.Printf("添加role_menus.tenant_id字段失败(可能已存在): %v", err)
	} else {
		log.Println("✅ role_menus.tenant_id字段检查/添加成功")
	}

	if err := db.Exec("ALTER TABLE role_apis ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 0").Error; err != nil {
		log.Printf("添加role_apis.tenant_id字段失败(可能已存在): %v", err)
	} else {
		log.Println("✅ role_apis.tenant_id字段检查/添加成功")
	}

	if err := db.Exec("ALTER TABLE user_roles ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 0").Error; err != nil {
		log.Printf("添加user_roles.tenant_id字段失败(可能已存在): %v", err)
	} else {
		log.Println("✅ user_roles.tenant_id字段检查/添加成功")
	}

	if err := db.Exec("ALTER TABLE tenant_roles ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 0").Error; err != nil {
		log.Printf("添加tenant_roles.tenant_id字段失败(可能已存在): %v", err)
	} else {
		log.Println("✅ tenant_roles.tenant_id字段检查/添加成功")
	}

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
		&model.TenantQuota{},
		&model.InstalledPlugin{},
		&model.PluginConfig{},
		&model.PluginEventLog{},
		&model.SSOAuthorizationCode{},
		&model.SSOTokenData{},
		&model.PaymentConfig{},
		&model.PaymentOrder{},
		&model.WebhookSubscription{},
		&model.WebhookDelivery{},
		&model.PluginVersionHistory{},
		&model.PluginDependency{},
		&model.AuditLog{},
		&model.Notification{},
		&model.TokenBlacklistEntry{},
		&model.FileRecord{},
		&model.APIKey{},
		&model.SettlementRecord{},
		&model.SettlementConfig{},
		&model.BackupRecord{},
		&model.Department{},
		&model.UserDepartment{},
		&model.RoleDept{},
		&model.LoginLog{},
		&model.DictType{},
		&model.DictData{},
		&model.ErrorCode{},
		&model.TenantPackage{},
		&model.TenantPackageMenu{},
		&model.CronJob{},
		&model.CronJobLog{},
		&model.Subscription{},
		&model.SubscriptionInvoice{},
		&model.NotificationChannel{},
		&model.NotificationTemplate{},
		&model.NotificationRecord{},
		&model.SensitiveWord{},
		&model.TenantChannelConfig{},
		&model.UserThirdParty{},
	}

	for _, m := range models {
		tableName := ""
		if t, ok := m.(interface{ TableName() string }); ok {
			tableName = t.TableName()
		}
		log.Printf("正在迁移表: %s (%T)...", tableName, m)

		if err := db.AutoMigrate(m); err != nil {
			log.Printf("❌ 迁移失败 %s (%T): %v", tableName, m, err)
			return fmt.Errorf("迁移失败 %T: %w", m, err)
		}
		log.Printf("✅ 迁移完成: %s (%T)", tableName, m)
	}

	log.Println("✅ 数据库迁移完成")

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id)").Error; err != nil {
		log.Printf("创建users.tenant_id索引失败: %v", err)
	} else {
		log.Println("✅ users.tenant_id索引创建成功")
	}

	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_tenant_username ON users(tenant_id, username)").Error; err != nil {
		log.Printf("创建复合唯一索引失败(可能已存在): %v", err)
	} else {
		log.Println("复合唯一索引 idx_tenant_username 创建成功")
	}

	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_tenant_plugin ON installed_plugins(tenant_id, plugin_id)").Error; err != nil {
		log.Printf("创建复合唯一索引失败(可能已存在): %v", err)
	} else {
		log.Println("复合唯一索引 idx_tenant_plugin 创建成功")
	}

	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_tenant_plugin_key ON plugin_configs(tenant_id, plugin_id, config_key)").Error; err != nil {
		log.Printf("创建复合唯一索引失败(可能已存在): %v", err)
	} else {
		log.Println("复合唯一索引 idx_tenant_plugin_key 创建成功")
	}

	return nil
}

func InitTestData(db *gorm.DB) error {
	log.Println("开始初始化测试数据...")

	ctx := utils.SkipTenantIsolation(context.Background())

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tx.Exec("DELETE FROM user_roles WHERE 1=1")
		tx.Exec("DELETE FROM role_menus WHERE 1=1")
		tx.Exec("DELETE FROM role_apis WHERE 1=1")
		tx.Exec("DELETE FROM tenant_roles WHERE 1=1")
		tx.Exec("DELETE FROM users WHERE username IN ('admin', 'test_user')")
		tx.Exec("DELETE FROM roles WHERE name IN ('super_admin', 'platform_admin', 'tenant_admin', 'tenant_user')")
		tx.Exec("DELETE FROM tenants WHERE name = '测试租户'")

		adminPassword := os.Getenv("FAYHUB_DEFAULT_ADMIN_PASSWORD")
		if adminPassword == "" {
			var err error
			adminPassword, err = utils.GenerateRandomPassword(16)
			if err != nil {
				return fmt.Errorf("生成随机密码失败: %w", err)
			}
			log.Printf("⚠️  未设置 FAYHUB_DEFAULT_ADMIN_PASSWORD，已生成随机密码: %s（请妥善保存）", adminPassword)
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("密码加密失败: %w", err)
		}

		defaultTenant := &model.Tenant{
			Name:        "测试租户",
			Domain:      "test.fayhub.com",
			Description: "系统默认创建的测试租户",
			Status:      1,
		}
		if err := tx.Create(defaultTenant).Error; err != nil {
			return fmt.Errorf("创建默认租户失败: %w", err)
		}

		adminUser := &model.User{
			SnowflakeTenantModel: model.SnowflakeTenantModel{TenantID: 0},
			Username:             "admin",
			Password:             string(hashedPassword),
			Email:                "admin@fayhub.com",
			Status:               1,
			Role:                 "super_admin",
			RealName:             "超级管理员",
		}
		if err := tx.Create(adminUser).Error; err != nil {
			return fmt.Errorf("创建超级管理员失败: %w", err)
		}

		testUser := &model.User{
			SnowflakeTenantModel: model.SnowflakeTenantModel{TenantID: defaultTenant.ID},
			Username:             "test_user",
			Password:             string(hashedPassword),
			Email:                "test@fayhub.com",
			Phone:                "13800138001",
			Status:               1,
			Role:                 "tenant_admin",
			RealName:             "测试用户",
		}
		if err := tx.Create(testUser).Error; err != nil {
			return fmt.Errorf("创建测试用户失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("初始化测试数据失败: %w", err)
	}

	log.Println("测试数据初始化完成")
	return nil
}
