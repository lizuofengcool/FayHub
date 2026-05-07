package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server       ServerConfig       `yaml:"server"`
	Database     DatabaseConfig     `yaml:"database"`
	Redis        RedisConfig        `yaml:"redis"`
	JWT          JWTConfig          `yaml:"jwt"`
	Logging      LoggingConfig      `yaml:"logging"`
	MultiTenant  MultiTenantConfig  `yaml:"multi_tenant"`
	Security     SecurityConfig     `yaml:"security"`
	PluginEngine PluginEngineConfig `yaml:"plugin_engine"`
	PluginSign   PluginSignConfig   `yaml:"plugin_sign"`
	Domains      DomainsConfig      `yaml:"domains"`
	Payment      PaymentConfig      `yaml:"payment"`
	SSO          SSOConfig          `yaml:"sso"`
	Storage      StorageConfig      `yaml:"storage"`
	System       SystemConfig       `yaml:"system"`
	Backup       BackupConfig       `yaml:"backup"`
}

type SystemConfig struct {
	ServiceToken    string `yaml:"service_token"`
	SnowflakeNodeID int64  `yaml:"snowflake_node_id"`
}

type BackupConfig struct {
	Enabled       bool     `yaml:"enabled"`
	Schedule      string   `yaml:"schedule"`
	RetentionDays int      `yaml:"retention_days"`
	MaxBackups    int      `yaml:"max_backups"`
	BackupDir     string   `yaml:"backup_dir"`
	Compress      bool     `yaml:"compress"`
	IncludeTables []string `yaml:"include_tables"`
	ExcludeTables []string `yaml:"exclude_tables"`
}

type SSOConfig struct {
	Clients map[string]string `yaml:"clients"`
}

type StorageConfig struct {
	Driver       string `yaml:"driver"`
	LocalPath    string `yaml:"local_path"`
	MaxSizeMB    int    `yaml:"max_size_mb"`
	AllowedTypes string `yaml:"allowed_types"`
	S3Endpoint   string `yaml:"s3_endpoint"`
	S3Region     string `yaml:"s3_region"`
	S3Bucket     string `yaml:"s3_bucket"`
	S3AccessKey  string `yaml:"s3_access_key"`
	S3SecretKey  string `yaml:"s3_secret_key"`
	S3UseSSL     bool   `yaml:"s3_use_ssl"`
}

type PaymentConfig struct {
	NotifyBaseURL    string `yaml:"notify_base_url"`
	OrderExpireMin   int    `yaml:"order_expire_min"`
	WechatGatewayURL string `yaml:"wechat_gateway_url"`
	AlipayGatewayURL string `yaml:"alipay_gateway_url"`
	AlipaySandboxURL string `yaml:"alipay_sandbox_url"`
}

type DomainsConfig struct {
	AdminURL     string `yaml:"admin_url"`
	MarketURL    string `yaml:"market_url"`
	MarketAPIURL string `yaml:"market_api_url"`
	DevURL       string `yaml:"dev_url"`
	APIURL       string `yaml:"api_url"`
	SSOURL       string `yaml:"sso_url"`
	WWWURL       string `yaml:"www_url"`
}

type PluginEngineConfig struct {
	HTTPTimeoutSec int    `yaml:"http_timeout_sec"`
	DefaultIconURL string `yaml:"default_icon_url"`
}

type PluginSignConfig struct {
	PublicKeyPath string `yaml:"public_key_path"`
}

type RedisConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

type SecurityConfig struct {
	MaxLoginAttempts int `yaml:"max_login_attempts"`
	LockDurationMin  int `yaml:"lock_duration_min"`
}

type ServerConfig struct {
	Port                 int      `yaml:"port"`
	Mode                 string   `yaml:"mode"`
	CORSOrigins          []string `yaml:"cors_origins"`
	CORSAllowAll         bool     `yaml:"cors_allow_all"`
	CORSAllowCredentials bool     `yaml:"cors_allow_credentials"`
}

type DatabaseConfig struct {
	Type            string `yaml:"type"`
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	Charset         string `yaml:"charset"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime_sec"`
	ConnMaxIdleTime int    `yaml:"conn_max_idle_time_sec"`
}

type JWTConfig struct {
	Secret         string `yaml:"secret"`
	Expire         int    `yaml:"expire"`
	Issuer         string `yaml:"issuer"`
	Algorithm      string `yaml:"algorithm"`
	PrivateKeyPath string `yaml:"private_key_path"`
	PublicKeyPath  string `yaml:"public_key_path"`
}

type LoggingConfig struct {
	Level  string            `yaml:"level"`
	Format string            `yaml:"format"`
	Output string            `yaml:"output"`
	File   LoggingFileConfig `yaml:"file"`
}

type LoggingFileConfig struct {
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

type MultiTenantConfig struct {
	Mode string `yaml:"mode"`
}

var GlobalConfig *Config
var configFilePath string

func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		env := os.Getenv("FAYHUB_ENV")
		if env == "" {
			env = "dev"
		}
		configPath = fmt.Sprintf("config_%s.yaml", env)
	}

	configFilePath = configPath

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	expanded := expandConfigContent(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("解析YAML配置失败: %w", err)
	}

	setDefaults(&cfg)

	overrideFromEnv(&cfg)

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	GlobalConfig = &cfg
	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.JWT.Expire == 0 {
		cfg.JWT.Expire = 168
	}
	if cfg.JWT.Issuer == "" {
		cfg.JWT.Issuer = "fayhub"
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
	if cfg.Logging.Output == "" {
		cfg.Logging.Output = "stdout"
	}
	if cfg.MultiTenant.Mode == "" {
		cfg.MultiTenant.Mode = "shared"
	}
	if cfg.Security.MaxLoginAttempts == 0 {
		cfg.Security.MaxLoginAttempts = 5
	}
	if cfg.Security.LockDurationMin == 0 {
		cfg.Security.LockDurationMin = 15
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 10
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 100
	}
	if cfg.Database.ConnMaxLifetime == 0 {
		cfg.Database.ConnMaxLifetime = 3600
	}
	if cfg.Database.ConnMaxIdleTime == 0 {
		cfg.Database.ConnMaxIdleTime = 600
	}
	if cfg.PluginEngine.HTTPTimeoutSec == 0 {
		cfg.PluginEngine.HTTPTimeoutSec = 30
	}
	if cfg.Redis.Host == "" {
		cfg.Redis.Host = "localhost"
	}
	if cfg.Redis.Port == 0 {
		cfg.Redis.Port = 6379
	}
	if cfg.Redis.PoolSize == 0 {
		cfg.Redis.PoolSize = 10
	}
	if cfg.Domains.AdminURL == "" {
		cfg.Domains.AdminURL = "http://admin.fayhub.com"
	}
	if cfg.Domains.MarketURL == "" {
		cfg.Domains.MarketURL = "http://www.fayhub.com"
	}
	if cfg.Domains.DevURL == "" {
		cfg.Domains.DevURL = "http://dev.fayhub.com"
	}
	if cfg.Domains.APIURL == "" {
		cfg.Domains.APIURL = "http://api.fayhub.com"
	}
	if cfg.Domains.SSOURL == "" {
		cfg.Domains.SSOURL = "http://sso.fayhub.com"
	}
	if cfg.Domains.WWWURL == "" {
		cfg.Domains.WWWURL = "http://www.fayhub.com"
	}
	if cfg.Payment.OrderExpireMin == 0 {
		cfg.Payment.OrderExpireMin = 30
	}
	if cfg.Payment.WechatGatewayURL == "" {
		cfg.Payment.WechatGatewayURL = "https://api.mch.weixin.qq.com"
	}
	if cfg.Payment.AlipayGatewayURL == "" {
		cfg.Payment.AlipayGatewayURL = "https://openapi.alipay.com/gateway.do"
	}
	if cfg.Payment.AlipaySandboxURL == "" {
		cfg.Payment.AlipaySandboxURL = "https://openapi.alipaydev.com/gateway.do"
	}
	if cfg.PluginEngine.DefaultIconURL == "" {
		cfg.PluginEngine.DefaultIconURL = "https://api.dicebear.com/7.x/identicon/svg"
	}
	if cfg.Storage.Driver == "" {
		cfg.Storage.Driver = "local"
	}
	if cfg.Storage.LocalPath == "" {
		cfg.Storage.LocalPath = "./uploads"
	}
	if cfg.Storage.MaxSizeMB == 0 {
		cfg.Storage.MaxSizeMB = 10
	}
	if cfg.Storage.AllowedTypes == "" {
		cfg.Storage.AllowedTypes = "jpg,jpeg,png,gif,pdf,doc,docx,xls,xlsx,zip"
	}
	if cfg.Backup.Schedule == "" {
		cfg.Backup.Schedule = "0 2 * * *"
	}
	if cfg.Backup.RetentionDays == 0 {
		cfg.Backup.RetentionDays = 30
	}
	if cfg.Backup.MaxBackups == 0 {
		cfg.Backup.MaxBackups = 100
	}
	if cfg.Backup.BackupDir == "" {
		cfg.Backup.BackupDir = "./data/backups"
	}
}

func overrideFromEnv(cfg *Config) {
	if v := os.Getenv("FAYHUB_DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("FAYHUB_DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("FAYHUB_DB_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Database.Port)
	}
	if v := os.Getenv("FAYHUB_DB_USERNAME"); v != "" {
		cfg.Database.Username = v
	}
	if v := os.Getenv("FAYHUB_DB_NAME"); v != "" {
		cfg.Database.Database = v
	}
	if v := os.Getenv("FAYHUB_JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("FAYHUB_SERVER_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Server.Port)
	}
	if v := os.Getenv("FAYHUB_SERVER_MODE"); v != "" {
		cfg.Server.Mode = v
	}
	if v := os.Getenv("FAYHUB_SECURITY_MAX_LOGIN_ATTEMPTS"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Security.MaxLoginAttempts)
	}
	if v := os.Getenv("FAYHUB_SECURITY_LOCK_DURATION_MIN"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Security.LockDurationMin)
	}
	if v := os.Getenv("FAYHUB_REDIS_ENABLED"); v == "true" {
		cfg.Redis.Enabled = true
	} else if v == "false" {
		cfg.Redis.Enabled = false
	}
	if v := os.Getenv("FAYHUB_REDIS_HOST"); v != "" {
		cfg.Redis.Host = v
	}
	if v := os.Getenv("FAYHUB_REDIS_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Redis.Port)
	}
	if v := os.Getenv("FAYHUB_REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := os.Getenv("FAYHUB_SERVICE_TOKEN"); v != "" {
		cfg.System.ServiceToken = v
	}
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("服务端口必须为1-65535之间的整数")
	}
	if c.Server.Mode != "debug" && c.Server.Mode != "release" {
		return fmt.Errorf("运行模式必须为debug或release")
	}
	if c.Database.Host != "" {
		if c.Database.Port <= 0 || c.Database.Port > 65535 {
			return fmt.Errorf("数据库端口必须为1-65535之间的整数")
		}
		if c.Database.Username == "" {
			return fmt.Errorf("数据库用户名不能为空")
		}
		if c.Database.Database == "" {
			return fmt.Errorf("数据库名称不能为空")
		}
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT密钥不能为空")
	}
	if c.JWT.Expire <= 0 {
		return fmt.Errorf("Token过期时间必须大于0")
	}
	if c.MultiTenant.Mode != "shared" && c.MultiTenant.Mode != "isolated" {
		return fmt.Errorf("多租户模式必须为shared或isolated")
	}
	return nil
}

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

func (l *LoggingFileConfig) GetLogFilePath() string {
	if l.Path == "" {
		l.Path = "./logs"
	}
	if err := os.MkdirAll(l.Path, 0755); err != nil {
		return ""
	}
	return filepath.Join(l.Path, "fayhub.log")
}

func SaveConfig() error {
	if configFilePath == "" {
		return fmt.Errorf("配置文件路径未初始化")
	}
	if GlobalConfig == nil {
		return fmt.Errorf("全局配置未加载")
	}

	data, err := yaml.Marshal(GlobalConfig)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configFilePath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

func GetConfigFilePath() string {
	return configFilePath
}

var envPattern = regexp.MustCompile(`\$\{([^}:]+)(?::-([^}]*))?\}`)

func expandEnvWithDefaults(s string) string {
	return envPattern.ReplaceAllStringFunc(s, func(match string) string {
		parts := envPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		key := parts[1]
		defaultVal := ""
		if len(parts) >= 3 {
			defaultVal = parts[2]
		}
		if val, ok := os.LookupEnv(key); ok {
			return val
		}
		return defaultVal
	})
}

func expandConfigContent(content string) string {
	content = expandEnvWithDefaults(content)
	content = os.ExpandEnv(content)
	return content
}

func autoTypeConvert(cfg *Config) {
	if cfg.Database.Port == 0 {
		if v, err := strconv.Atoi(expandEnvWithDefaults("${FAYHUB_DB_PORT:-5432}")); err == nil {
			cfg.Database.Port = v
		}
	}
	if cfg.Server.Port == 0 {
		if v, err := strconv.Atoi(expandEnvWithDefaults("${FAYHUB_SERVER_PORT:-8080}")); err == nil {
			cfg.Server.Port = v
		}
	}
}
