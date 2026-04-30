package domains

import (
	"fmt"

	"fayhub/pkg/config"
)

type DomainConfig struct {
	Admin string `json:"admin"`
	API   string `json:"api"`
	WWW   string `json:"www"`
	Dev   string `json:"dev"`
	SSO   string `json:"sso"`
}

var Current *DomainConfig

func Init() {
	if config.GlobalConfig != nil {
		cfg := config.GlobalConfig.Domains
		Current = &DomainConfig{
			Admin: cfg.AdminURL,
			API:   cfg.APIURL,
			WWW:   cfg.WWWURL,
			Dev:   cfg.DevURL,
			SSO:   cfg.SSOURL,
		}
	}

	if Current == nil {
		Current = &DomainConfig{
			Admin: "http://admin.fayhub.com",
			API:   "http://api.fayhub.com",
			WWW:   "http://www.fayhub.com",
			Dev:   "http://dev.fayhub.com",
			SSO:   "http://sso.fayhub.com",
		}
	}
}

func GetAdminURL() string {
	if Current == nil {
		Init()
	}
	return Current.Admin
}

func GetAPIURL() string {
	if Current == nil {
		Init()
	}
	return Current.API
}

func GetMarketURL() string {
	if Current == nil {
		Init()
	}
	return Current.WWW
}

func GetWWWURL() string {
	if Current == nil {
		Init()
	}
	return Current.WWW
}

func GetDevURL() string {
	if Current == nil {
		Init()
	}
	return Current.Dev
}

func GetSSOURL() string {
	if Current == nil {
		Init()
	}
	return Current.SSO
}

func GetAllCORSOrigins() []string {
	if Current == nil {
		Init()
	}

	origins := []string{
		Current.Admin,
		Current.API,
		Current.WWW,
		Current.Dev,
		Current.SSO,
	}

	if config.GlobalConfig != nil {
		for _, o := range config.GlobalConfig.Server.CORSOrigins {
			origins = append(origins, o)
		}
	}

	if config.GlobalConfig == nil || config.GlobalConfig.Server.Mode == "debug" {
		origins = append(origins,
			"http://localhost:3000",
			"http://localhost:3002",
			"http://localhost:3003",
			"http://localhost:5173",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3002",
			"http://127.0.0.1:3003",
			"http://127.0.0.1:5173",
		)
	}

	return origins
}

func Validate() error {
	if Current == nil {
		return fmt.Errorf("域名配置未初始化")
	}

	required := []struct {
		name  string
		value string
	}{
		{"Admin", Current.Admin},
		{"API", Current.API},
		{"WWW", Current.WWW},
		{"Dev", Current.Dev},
		{"SSO", Current.SSO},
	}

	for _, field := range required {
		if field.value == "" {
			return fmt.Errorf("域名配置缺失: %s", field.name)
		}
	}

	return nil
}

func init() {
	Init()
}
