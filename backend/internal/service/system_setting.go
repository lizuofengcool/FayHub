package service

import (
	"context"
	"sync"

	"fayhub/pkg/config"
	"fayhub/pkg/domains"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/utils"
)

type SystemSettingService struct{}

var (
	runtimeSettings map[string]interface{}
	settingsMu      sync.RWMutex
)

func init() {
	runtimeSettings = make(map[string]interface{})
}

type DomainSettings struct {
	AdminURL  string `json:"admin_url"`
	MarketURL string `json:"market_url"`
	DevURL    string `json:"dev_url"`
	APIURL    string `json:"api_url"`
	SSOURL    string `json:"sso_url"`
	WWWURL    string `json:"www_url"`
}

type PaymentSettings struct {
	NotifyBaseURL    string `json:"notify_base_url"`
	OrderExpireMin   int    `json:"order_expire_min"`
	WechatGatewayURL string `json:"wechat_gateway_url"`
	AlipayGatewayURL string `json:"alipay_gateway_url"`
	AlipaySandboxURL string `json:"alipay_sandbox_url"`
}

type SecuritySettings struct {
	MaxLoginAttempts int `json:"max_login_attempts"`
	LockDurationMin  int `json:"lock_duration_min"`
}

type SystemSettingsResponse struct {
	Domains  DomainSettings   `json:"domains"`
	Payment  PaymentSettings  `json:"payment"`
	Security SecuritySettings `json:"security"`
	Server   ServerSettings   `json:"server"`
	Backup   BackupSettings   `json:"backup"`
}

type BackupSettings struct {
	Enabled       bool     `json:"enabled"`
	Schedule      string   `json:"schedule"`
	RetentionDays int      `json:"retention_days"`
	MaxBackups    int      `json:"max_backups"`
	BackupDir     string   `json:"backup_dir"`
	Compress      bool     `json:"compress"`
	IncludeTables []string `json:"include_tables"`
	ExcludeTables []string `json:"exclude_tables"`
}

type ServerSettings struct {
	Port int    `json:"port"`
	Mode string `json:"mode"`
}

type UpdateSettingsRequest struct {
	Domains  *DomainSettings   `json:"domains"`
	Payment  *PaymentSettings  `json:"payment"`
	Security *SecuritySettings `json:"security"`
	Backup   *BackupSettings   `json:"backup"`
}

func (s *SystemSettingService) GetSettings(ctx context.Context) (*SystemSettingsResponse, error) {
	cfg := config.GlobalConfig
	if cfg == nil {
		return nil, errs.NewServiceError(errs.ErrConfigNotLoaded, "")
	}

	return &SystemSettingsResponse{
		Domains: DomainSettings{
			AdminURL:  cfg.Domains.AdminURL,
			MarketURL: cfg.Domains.MarketURL,
			DevURL:    cfg.Domains.DevURL,
			APIURL:    cfg.Domains.APIURL,
			SSOURL:    cfg.Domains.SSOURL,
			WWWURL:    cfg.Domains.WWWURL,
		},
		Payment: PaymentSettings{
			NotifyBaseURL:    cfg.Payment.NotifyBaseURL,
			OrderExpireMin:   cfg.Payment.OrderExpireMin,
			WechatGatewayURL: cfg.Payment.WechatGatewayURL,
			AlipayGatewayURL: cfg.Payment.AlipayGatewayURL,
			AlipaySandboxURL: cfg.Payment.AlipaySandboxURL,
		},
		Security: SecuritySettings{
			MaxLoginAttempts: cfg.Security.MaxLoginAttempts,
			LockDurationMin:  cfg.Security.LockDurationMin,
		},
		Server: ServerSettings{
			Port: cfg.Server.Port,
			Mode: cfg.Server.Mode,
		},
		Backup: BackupSettings{
			Enabled:       cfg.Backup.Enabled,
			Schedule:      cfg.Backup.Schedule,
			RetentionDays: cfg.Backup.RetentionDays,
			MaxBackups:    cfg.Backup.MaxBackups,
			BackupDir:     cfg.Backup.BackupDir,
			Compress:      cfg.Backup.Compress,
			IncludeTables: cfg.Backup.IncludeTables,
			ExcludeTables: cfg.Backup.ExcludeTables,
		},
	}, nil
}

func (s *SystemSettingService) UpdateSettings(ctx context.Context, req UpdateSettingsRequest) error {
	cfg := config.GlobalConfig
	if cfg == nil {
		return errs.NewServiceError(errs.ErrConfigNotLoaded, "系统配置未加载")
	}

	if req.Domains != nil {
		if req.Domains.AdminURL != "" {
			cfg.Domains.AdminURL = req.Domains.AdminURL
		}
		if req.Domains.MarketURL != "" {
			cfg.Domains.MarketURL = req.Domains.MarketURL
		}
		if req.Domains.DevURL != "" {
			cfg.Domains.DevURL = req.Domains.DevURL
		}
		if req.Domains.APIURL != "" {
			cfg.Domains.APIURL = req.Domains.APIURL
		}
		if req.Domains.SSOURL != "" {
			cfg.Domains.SSOURL = req.Domains.SSOURL
		}
		if req.Domains.WWWURL != "" {
			cfg.Domains.WWWURL = req.Domains.WWWURL
		}

		syncDomainsPackage()
	}

	if req.Payment != nil {
		if req.Payment.NotifyBaseURL != "" {
			cfg.Payment.NotifyBaseURL = req.Payment.NotifyBaseURL
		}
		if req.Payment.OrderExpireMin > 0 {
			cfg.Payment.OrderExpireMin = req.Payment.OrderExpireMin
		}
		if req.Payment.WechatGatewayURL != "" {
			cfg.Payment.WechatGatewayURL = req.Payment.WechatGatewayURL
		}
		if req.Payment.AlipayGatewayURL != "" {
			cfg.Payment.AlipayGatewayURL = req.Payment.AlipayGatewayURL
		}
		if req.Payment.AlipaySandboxURL != "" {
			cfg.Payment.AlipaySandboxURL = req.Payment.AlipaySandboxURL
		}
	}

	if req.Security != nil {
		if req.Security.MaxLoginAttempts > 0 {
			cfg.Security.MaxLoginAttempts = req.Security.MaxLoginAttempts
		}
		if req.Security.LockDurationMin > 0 {
			cfg.Security.LockDurationMin = req.Security.LockDurationMin
		}
	}

	if req.Backup != nil {
		cfg.Backup.Enabled = req.Backup.Enabled
		if req.Backup.Schedule != "" {
			cfg.Backup.Schedule = req.Backup.Schedule
		}
		if req.Backup.RetentionDays > 0 {
			cfg.Backup.RetentionDays = req.Backup.RetentionDays
		}
		if req.Backup.MaxBackups > 0 {
			cfg.Backup.MaxBackups = req.Backup.MaxBackups
		}
		if req.Backup.BackupDir != "" {
			cfg.Backup.BackupDir = req.Backup.BackupDir
		}
		cfg.Backup.Compress = req.Backup.Compress
		if req.Backup.IncludeTables != nil {
			cfg.Backup.IncludeTables = req.Backup.IncludeTables
		}
		if req.Backup.ExcludeTables != nil {
			cfg.Backup.ExcludeTables = req.Backup.ExcludeTables
		}
	}

	settingsMu.Lock()
	userID, _ := utils.GetUserIDFromContext(ctx)
	runtimeSettings["last_updated_by"] = userID
	settingsMu.Unlock()

	if err := config.SaveConfig(); err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "运行时配置已更新，但持久化失败")
	}

	return nil
}

func syncDomainsPackage() {
	cfg := config.GlobalConfig
	if cfg == nil {
		return
	}

	domains.Current = &domains.DomainConfig{
		Admin: cfg.Domains.AdminURL,
		API:   cfg.Domains.APIURL,
		WWW:   cfg.Domains.WWWURL,
		Dev:   cfg.Domains.DevURL,
		SSO:   cfg.Domains.SSOURL,
	}
}
