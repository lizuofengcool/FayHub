package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config 全局配置结构体
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	JWT         JWTConfig         `yaml:"jwt"`
	Logging     LoggingConfig     `yaml:"logging"`
	MultiTenant MultiTenantConfig `yaml:"multi_tenant"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Charset  string `yaml:"charset"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string `yaml:"secret"`
	Expire int    `yaml:"expire"`
	Issuer string `yaml:"issuer"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string         `yaml:"level"`
	Format string         `yaml:"format"`
	Output string         `yaml:"output"`
	File   LoggingFileConfig `yaml:"file"`
}

// LoggingFileConfig 日志文件配置
type LoggingFileConfig struct {
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

// MultiTenantConfig 多租户配置
type MultiTenantConfig struct {
	Mode string `yaml:"mode"` // shared: 共享库, isolated: 独立库
}

// ConfigManager 配置管理器
type ConfigManager struct {
	config *Config
	mu     sync.RWMutex
}

var (
	globalConfig *Config
	configOnce   sync.Once
)

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	var err error
	configOnce.Do(func() {
		globalConfig, err = loadConfigFromFile(configPath)
		if err != nil {
			log.Fatalf("加载配置文件失败: %v", err)
		}
		
		// 验证配置
		if err := globalConfig.Validate(); err != nil {
			log.Fatalf("配置验证失败: %v", err)
		}
		
		log.Printf("配置文件加载成功: %s", configPath)
	})
	
	return globalConfig, err
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	if globalConfig == nil {
		log.Fatal("配置未初始化，请先调用 LoadConfig")
	}
	return globalConfig
}

// loadConfigFromFile 从文件加载配置
func loadConfigFromFile(configPath string) (*Config, error) {
	// 如果未指定配置文件路径，使用默认路径
	if configPath == "" {
		env := os.Getenv("FAYHUB_ENV")
		if env == "" {
			env = "dev"
		}
		configPath = fmt.Sprintf("config_%s.yaml", env)
	}
	
	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}
	
	// 读取文件内容
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}
	
	// 解析YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析YAML配置失败: %v", err)
	}
	
	return &config, nil
}

// Validate 配置验证
func (c *Config) Validate() error {
	// 服务配置验证
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("服务端口必须为1-65535之间的整数")
	}
	if c.Server.Mode != "debug" && c.Server.Mode != "release" {
		return fmt.Errorf("运行模式必须为debug或release")
	}
	
	// 数据库配置验证
	if c.Database.Host == "" {
		return fmt.Errorf("数据库地址不能为空")
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		return fmt.Errorf("数据库端口必须为1-65535之间的整数")
	}
	if c.Database.Username == "" {
		return fmt.Errorf("数据库用户名不能为空")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("数据库名称不能为空")
	}
	
	// JWT配置验证
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT密钥不能为空")
	}
	if c.JWT.Expire <= 0 {
		return fmt.Errorf("Token过期时间必须大于0")
	}
	
	// 多租户配置验证
	if c.MultiTenant.Mode == "" {
		return fmt.Errorf("多租户模式不能为空")
	}
	if c.MultiTenant.Mode != "shared" && c.MultiTenant.Mode != "isolated" {
		return fmt.Errorf("多租户模式必须为shared或isolated")
	}
	
	// 日志配置验证
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
	if c.Logging.Output == "" {
		c.Logging.Output = "stdout"
	}
	
	return nil
}

// GetDSN 获取数据库连接字符串
func (d *DatabaseConfig) GetDSN() string {
	switch d.Type {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			d.Username, d.Password, d.Host, d.Port, d.Database, d.Charset)
	case "postgresql":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			d.Host, d.Port, d.Username, d.Password, d.Database)
	case "sqlite":
		return d.Database
	default:
		return ""
	}
}

// GetLogFilePath 获取日志文件完整路径
func (l *LoggingFileConfig) GetLogFilePath() string {
	if l.Path == "" {
		l.Path = "./logs"
	}
	
	// 创建日志目录
	if err := os.MkdirAll(l.Path, 0755); err != nil {
		log.Printf("创建日志目录失败: %v", err)
		return ""
	}
	
	return filepath.Join(l.Path, "fayhub.log")
}