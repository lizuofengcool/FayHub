package service

import (
	"context"
	"fmt"
)

// ServiceGroup 服务组管理（GVA 标准工程实践）
// 作用：统一管理所有业务服务实例，避免零散初始化
type ServiceGroup struct {
	// 系统核心服务
	SystemService
	// 认证服务（阶段二新增）
	AuthService
	// 预留扩展：后续可添加 TenantService、UserService 等
	// TenantService
	// UserService
}

// 实例化全局服务组（对外暴露，供控制器/路由调用）
var ServiceGroupApp = new(ServiceGroup)

// ==================== 系统服务子组（阶段一核心）====================
// SystemService 系统基础服务
// 包含：健康检查、系统配置等基础能力
type SystemService struct{}

// 初始化系统服务（空实现，阶段一先搭结构，后续补充健康检查逻辑）
func (s *SystemService) Init() {}

// HealthCheck 健康检查核心方法（阶段一关键）
// 作用：从上下文提取租户ID并返回，验证租户隔离能力
// 参数：ctx 上下文（包含租户ID）
// 返回：租户ID、提示信息、错误
func (s *SystemService) HealthCheck(ctx context.Context) (uint, string, error) {
	// 从上下文中提取租户ID
	tenantIDValue := ctx.Value("tenant_id")
	if tenantIDValue == nil {
		// 如果没有租户ID，说明是总后台管理员操作
		return 0, "系统运行正常（总后台管理员）", nil
	}
	
	// 类型断言获取租户ID
	tentantID, ok := tenantIDValue.(uint)
	if !ok {
		// 类型转换失败，返回错误
		return 0, "", fmt.Errorf("租户ID类型错误")
	}
	
	if tentantID == 0 {
		return 0, "系统运行正常（总后台管理员）", nil
	}
	
	return tentantID, fmt.Sprintf("系统运行正常（租户ID：%d）", tentantID), nil
}