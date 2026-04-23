package config

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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