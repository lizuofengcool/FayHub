package service

import (
	"context"
	"fmt"
)

type ServiceGroup struct {
	SystemService
	AuthService
	TenantService
	UserService
}

var ServiceGroupApp = new(ServiceGroup)

type SystemService struct{}

func (s *SystemService) Init() {}

func (s *SystemService) HealthCheck(ctx context.Context) (uint, string, error) {
	tenantIDValue := ctx.Value("tenant_id")
	if tenantIDValue == nil {
		return 0, "系统运行正常（总后台管理员）", nil
	}

	tentantID, ok := tenantIDValue.(uint)
	if !ok {
		return 0, "", fmt.Errorf("租户ID类型错误")
	}

	if tentantID == 0 {
		return 0, "系统运行正常（总后台管理员）", nil
	}

	return tentantID, fmt.Sprintf("系统运行正常（租户ID：%d）", tentantID), nil
}
