// +build ignore

package main

import (
	"fayhub/pkg/config"
	"fayhub/internal/initialize"
	"fayhub/internal/model"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

func main() {
	log.Println("🧪 开始数据库连接测试...")

	// 加载配置文件
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("❌ 加载配置文件失败: %v", err)
	}

	log.Println("✅ 配置文件加载成功")

	// 测试数据库连接
	log.Println("🔗 测试数据库连接...")
	db, err := initialize.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	log.Println("✅ 数据库连接成功")

	// 测试数据库操作
	log.Println("📊 测试数据库操作...")
	testDatabaseOperations(db)

	log.Println("🎉 数据库测试完成！")
}

func testDatabaseOperations(db *gorm.DB) {
	// 测试租户表操作
	tenant := &model.Tenant{
		Name:        "测试租户",
		Code:        "TEST",
		ContactName: "测试联系人",
		ContactPhone: "13800000000",
		Status:      1,
	}

	if err := db.Create(tenant).Error; err != nil {
		log.Printf("⚠️  创建租户失败: %v", err)
	} else {
		log.Printf("✅ 创建租户成功: ID=%d, 名称=%s", tenant.ID, tenant.Name)
	}

	// 查询租户列表
	var tenants []model.Tenant
	if err := db.Find(&tenants).Error; err != nil {
		log.Printf("⚠️  查询租户列表失败: %v", err)
	} else {
		log.Printf("✅ 查询到 %d 个租户", len(tenants))
	}

	// 测试用户表操作
	user := &model.User{
		Username: "testuser",
		Nickname: "测试用户",
		Password: "test123",
		TenantID: tenant.ID,
		Status:   1,
	}

	if err := db.Create(user).Error; err != nil {
		log.Printf("⚠️  创建用户失败: %v", err)
	} else {
		log.Printf("✅ 创建用户成功: ID=%d, 用户名=%s", user.ID, user.Username)
	}

	// 查询用户列表
	var users []model.User
	if err := db.Find(&users).Error; err != nil {
		log.Printf("⚠️  查询用户列表失败: %v", err)
	} else {
		log.Printf("✅ 查询到 %d 个用户", len(users))
	}

	// 清理测试数据
	db.Delete(user)
	db.Delete(tenant)
	log.Println("🧹 测试数据清理完成")
}