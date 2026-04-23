package initialize

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string `yaml:"type"`     // 数据库类型（mysql/postgresql）
	Host     string `yaml:"host"`     // 数据库地址
	Port     int    `yaml:"port"`     // 数据库端口
	Username string `yaml:"username"` // 数据库用户名
	Password string `yaml:"password"` // 数据库密码
	Database string `yaml:"database"` // 数据库名称
	Charset  string `yaml:"charset"`  // 数据库字符集
}

// InitDB 初始化数据库连接
func InitDB(config *DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch config.Type {
	case "mysql":
		// MySQL DSN格式：username:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			config.Username, config.Password, config.Host, config.Port, config.Database, config.Charset)
		dialector = mysql.Open(dsn)
	case "postgresql":
		// PostgreSQL DSN格式：host=host port=port user=user password=password dbname=dbname sslmode=disable TimeZone=Asia/Shanghai
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
			config.Host, config.Port, config.Username, config.Password, config.Database)
		dialector = postgres.Open(dsn)
	case "sqlite":
		// SQLite DSN格式：file:test.db?cache=shared&mode=memory
		dialector = sqlite.Open(config.Database)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	// 初始化GORM DB实例
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 设置日志级别
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %v", err)
	}

	return db, nil
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate(db *gorm.DB, models ...interface{}) error {
	return db.AutoMigrate(models...)
}