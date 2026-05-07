package service

import (
	"context"
	"fmt"

	"fayhub/pkg/utils"
)

type ServiceGroup struct {
	SystemService
	SystemSettingService
	AuthService
	TenantService
	UserService
	RBACService
	MenuService
	APIService
	PluginEngineService
	PluginResourceMonitorService
	SSOService
	LicenseService
	CaptchaService
	PaymentService
	FileService
	OnlineUserService
	CronJobService
	SubscriptionService
	NotificationChannelService
	StatsService
	SensitiveWordService
	ExcelService
}

var ServiceGroupApp = new(ServiceGroup)

type SystemService struct{}

func (s *SystemService) Init() {}

func (s *SystemService) HealthCheck(ctx context.Context) (int64, string, error) {
	tenantID, ok := utils.GetTenantIDFromCtx(ctx)
	if !ok || tenantID == 0 {
		return 0, "系统运行正常（总后台管理员）", nil
	}

	return tenantID, fmt.Sprintf("系统运行正常（租户ID：%d）", tenantID), nil
}
