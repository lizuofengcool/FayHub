package service

import "context"

// ServiceGroup 服务组管理（GVA 标准工程实践）
// 作用：统一管理所有业务服务实例，避免零散初始化
type ServiceGroup struct {
	// 系统核心服务（阶段一先搭骨架，后续填充健康检查/租户管理逻辑）
	SystemService
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
	// 阶段一先返回骨架，后续结合中间件上下文完善逻辑
	// 占位：从 ctx 提取 X-Tenant-ID
	var tenantID uint = 0
	return tenantID, "系统运行正常", nil
}