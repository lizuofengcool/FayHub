package errors

// 通用错误码 (40000-40999)
const (
	// 通用错误码 (40000-40999)
	ErrParamValidation  = 40000 // 参数校验失败
	ErrParamMissing     = 40001 // 请求参数缺失
	ErrParamFormat      = 40002 // 请求参数格式错误
	ErrResourceNotFound = 40003 // 资源不存在
	ErrOperationFailed  = 40004 // 操作失败
	ErrConflict         = 40009 // 资源冲突

	// 认证错误码 (41000-41999)
	ErrUnauthorized      = 41000 // 未登录
	ErrTokenExpired      = 41001 // Token过期
	ErrTokenInvalid      = 41002 // Token无效
	ErrLoginFailed       = 41003 // 登录失败
	ErrPasswordIncorrect = 41004 // 密码错误
	ErrAccountDisabled   = 41005 // 账号已禁用
	ErrTokenRevoked      = 41006 // Token已注销

	// 权限错误码 (42000-42999)
	ErrPermissionDenied = 42000 // 无权限访问
	ErrTenantPermission = 42001 // 租户权限不足
	ErrDataPermission   = 42002 // 数据权限不足
	ErrRoleNotExist     = 42003 // 角色不存在
	ErrMenuNotExist     = 42004 // 菜单不存在

	// 租户隔离错误码 (42100-42199)
	ErrTenantIDMissing      = 42100 // 租户ID缺失
	ErrCrossTenantOperation = 42101 // 跨租户操作被禁止
	ErrTenantNotExist       = 42102 // 租户不存在或已禁用
	ErrTenantDisabled       = 42103 // 租户已禁用
	ErrTenantQuotaExceeded  = 42104 // 租户配额已满

	// 业务错误码 (43000-49999)
	ErrUserNotExist       = 43000 // 用户不存在
	ErrUserAlreadyExist   = 43001 // 用户已存在
	ErrEmailAlreadyUsed   = 43002 // 邮箱已被使用
	ErrPhoneAlreadyUsed   = 43003 // 手机号已被使用
	ErrDepartmentNotExist = 43004 // 部门不存在
	ErrPositionNotExist   = 43005 // 职位不存在
	ErrPasswordWeak       = 43006 // 密码强度不足
	ErrOldPasswordWrong   = 43007 // 原密码错误
	ErrAccountLocked      = 43008 // 账户已锁定
	ErrDBNotConnected     = 43009 // 数据库未连接
	ErrRoleAlreadyBound   = 43010 // 角色已绑定
	ErrRoleNotBound       = 43011 // 角色未绑定
	ErrRoleAlreadyExist   = 43012 // 角色名称已存在
	ErrForbidden          = 43013 // 操作被禁止
	ErrAPINotExist        = 43014 // API接口不存在
	ErrAPIAlreadyExist    = 43015 // API接口已存在
	ErrAPINotInstalled    = 43016 // API未安装
	ErrCaptchaFailed      = 43017 // 验证码生成失败
	ErrLicenseInvalid     = 43018 // License格式无效
	ErrSSOCodeInvalid     = 43019 // SSO授权码无效
	ErrSSOCodeExpired     = 43020 // SSO授权码已过期
	ErrSSONotLoggedIn     = 43021 // SSO用户未登录
	ErrSSOCodeGenerate    = 43022 // SSO授权码生成失败

	// 支付错误码 (43100-43199)
	ErrPaymentConfigNotFound  = 43100 // 支付配置不存在
	ErrPaymentChannelDisabled = 43101 // 支付渠道未启用
	ErrPaymentOrderNotFound   = 43102 // 支付订单不存在
	ErrPaymentOrderExpired    = 43103 // 支付订单已过期
	ErrPaymentOrderPaid       = 43104 // 订单已支付
	ErrPaymentAmountInvalid   = 43105 // 支付金额无效
	ErrPaymentSignFailed      = 43106 // 支付签名失败
	ErrPaymentNotifyFailed    = 43107 // 支付回调处理失败
	ErrPaymentRefundFailed    = 43108 // 退款失败

	// 应用市场错误码 (44000-44999)
	ErrPluginNotFound         = 44000 // 插件不存在
	ErrPluginNotApproved      = 44001 // 插件未上架
	ErrPluginAlreadyInstalled = 44002 // 插件已安装
	ErrPluginNotInstalled     = 44003 // 插件未安装
	ErrPluginInstallFailed    = 44004 // 插件安装失败
	ErrPluginUninstallFailed  = 44005 // 插件卸载失败
	ErrPluginVersionConflict  = 44006 // 插件版本冲突
	ErrPluginDisabled         = 44007 // 插件已禁用
	ErrPluginVersionNotFound  = 44008 // 插件版本不存在
	ErrDeveloperNotFound      = 44010 // 开发者不存在
	ErrDeveloperPending       = 44011 // 开发者待审核
	ErrDeveloperSuspended     = 44012 // 开发者已暂停
	ErrDeveloperAlreadyExist  = 44013 // 开发者已存在
	ErrCategoryNotFound       = 44020 // 分类不存在
	ErrCategoryAlreadyExist   = 44021 // 分类已存在
	ErrTagNotFound            = 44030 // 标签不存在
	ErrTagAlreadyExist        = 44031 // 标签已存在
	ErrSlugAlreadyExist       = 44040 // 插件标识已存在

	// 系统错误码 (50000-59999)
	ErrInternalServer  = 50000 // 服务器内部错误
	ErrDatabase        = 50001 // 数据库错误
	ErrCache           = 50002 // 缓存错误
	ErrFileSystem      = 50003 // 文件系统错误
	ErrNetwork         = 50004 // 网络错误
	ErrExternalService = 50005 // 外部服务错误
	ErrConfigNotLoaded = 50006 // 系统配置未加载
)

// ErrorMessages 错误码映射中文描述
var ErrorMessages = map[int]string{
	// 通用错误码
	ErrParamValidation:  "参数校验失败",
	ErrParamMissing:     "请求参数缺失",
	ErrParamFormat:      "请求参数格式错误",
	ErrResourceNotFound: "资源不存在",
	ErrOperationFailed:  "操作失败",
	ErrConflict:         "资源冲突",

	// 认证错误码
	ErrUnauthorized:      "未登录",
	ErrTokenExpired:      "Token已过期",
	ErrTokenInvalid:      "Token无效",
	ErrLoginFailed:       "登录失败",
	ErrPasswordIncorrect: "密码错误",
	ErrAccountDisabled:   "用户已被禁用",
	ErrTokenRevoked:      "Token已注销",

	// 权限错误码
	ErrPermissionDenied: "无权限访问",
	ErrTenantPermission: "租户权限不足",
	ErrDataPermission:   "数据权限不足",
	ErrRoleNotExist:     "角色不存在",
	ErrMenuNotExist:     "菜单不存在",

	// 租户隔离错误码
	ErrTenantIDMissing:      "租户ID缺失",
	ErrCrossTenantOperation: "跨租户操作被禁止",
	ErrTenantNotExist:       "租户不存在或已禁用",
	ErrTenantDisabled:       "租户已禁用",
	ErrTenantQuotaExceeded:  "租户配额已满",

	// 业务错误码
	ErrUserNotExist:       "用户不存在",
	ErrUserAlreadyExist:   "用户名已存在",
	ErrEmailAlreadyUsed:   "邮箱已被使用",
	ErrPhoneAlreadyUsed:   "手机号已被使用",
	ErrDepartmentNotExist: "部门不存在",
	ErrPositionNotExist:   "职位不存在",
	ErrPasswordWeak:       "密码强度不足",
	ErrOldPasswordWrong:   "原密码错误",
	ErrAccountLocked:      "账户已锁定",
	ErrDBNotConnected:     "数据库未连接",
	ErrRoleAlreadyBound:   "角色已绑定",
	ErrRoleNotBound:       "角色未绑定",
	ErrRoleAlreadyExist:   "角色名称已存在",
	ErrForbidden:          "操作被禁止",
	ErrAPINotExist:        "API接口不存在",
	ErrAPIAlreadyExist:    "API接口已存在",
	ErrAPINotInstalled:    "API未安装",
	ErrCaptchaFailed:      "验证码生成失败",
	ErrLicenseInvalid:     "License格式无效",
	ErrSSOCodeInvalid:     "授权码无效",
	ErrSSOCodeExpired:     "授权码已过期",
	ErrSSONotLoggedIn:     "用户未登录",
	ErrSSOCodeGenerate:    "授权码生成失败",

	ErrPaymentConfigNotFound:  "支付配置不存在",
	ErrPaymentChannelDisabled: "支付渠道未启用",
	ErrPaymentOrderNotFound:   "支付订单不存在",
	ErrPaymentOrderExpired:    "支付订单已过期",
	ErrPaymentOrderPaid:       "订单已支付",
	ErrPaymentAmountInvalid:   "支付金额无效",
	ErrPaymentSignFailed:      "支付签名失败",
	ErrPaymentNotifyFailed:    "支付回调处理失败",
	ErrPaymentRefundFailed:    "退款失败",

	// 应用市场错误码
	ErrPluginNotFound:         "插件不存在",
	ErrPluginNotApproved:      "插件未上架",
	ErrPluginAlreadyInstalled: "插件已安装",
	ErrPluginNotInstalled:     "插件未安装",
	ErrPluginInstallFailed:    "插件安装失败",
	ErrPluginUninstallFailed:  "插件卸载失败",
	ErrPluginVersionConflict:  "插件版本冲突",
	ErrPluginDisabled:         "插件已禁用",
	ErrPluginVersionNotFound:  "插件版本不存在",
	ErrDeveloperNotFound:      "开发者不存在",
	ErrDeveloperPending:       "开发者待审核",
	ErrDeveloperSuspended:     "开发者已暂停",
	ErrDeveloperAlreadyExist:  "开发者已存在",
	ErrCategoryNotFound:       "分类不存在",
	ErrCategoryAlreadyExist:   "分类已存在",
	ErrTagNotFound:            "标签不存在",
	ErrTagAlreadyExist:        "标签已存在",
	ErrSlugAlreadyExist:       "插件标识已存在",

	// 系统错误码
	ErrInternalServer:  "服务器内部错误",
	ErrDatabase:        "数据库错误",
	ErrCache:           "缓存错误",
	ErrFileSystem:      "文件系统错误",
	ErrNetwork:         "网络错误",
	ErrExternalService: "外部服务错误",
	ErrConfigNotLoaded: "系统配置未加载",
}

// GetErrorMessage 获取错误码对应的中文描述
func GetErrorMessage(code int) string {
	if msg, exists := ErrorMessages[code]; exists {
		return msg
	}
	return "未知错误"
}

// IsClientError 判断是否为客户端错误
func IsClientError(code int) bool {
	return code >= 40000 && code < 50000
}

// IsServerError 判断是否为服务端错误
func IsServerError(code int) bool {
	return code >= 50000 && code < 60000
}

// IsAuthError 判断是否为认证错误
func IsAuthError(code int) bool {
	return code >= 41000 && code < 42000
}

// IsPermissionError 判断是否为权限错误
func IsPermissionError(code int) bool {
	return code >= 42000 && code < 43000
}

// IsTenantError 判断是否为租户错误
func IsTenantError(code int) bool {
	return code >= 42100 && code < 42200
}

// IsBusinessError 判断是否为业务错误
func IsBusinessError(code int) bool {
	return code >= 43000 && code < 50000
}

// ErrorCode 错误码结构体
type ErrorCode struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewErrorCode 创建错误码实例
func NewErrorCode(code int) *ErrorCode {
	return &ErrorCode{
		Code:    code,
		Message: GetErrorMessage(code),
	}
}

// WithMessage 自定义错误消息
func (ec *ErrorCode) WithMessage(message string) *ErrorCode {
	ec.Message = message
	return ec
}

// ToResponse 转换为响应结构
func (ec *ErrorCode) ToResponse() map[string]interface{} {
	return map[string]interface{}{
		"code":    ec.Code,
		"message": ec.Message,
	}
}

// 预定义错误码实例
var (
	// 通用错误
	ParamValidationError  = NewErrorCode(ErrParamValidation)
	ParamMissingError     = NewErrorCode(ErrParamMissing)
	ParamFormatError      = NewErrorCode(ErrParamFormat)
	ResourceNotFoundError = NewErrorCode(ErrResourceNotFound)
	OperationFailedError  = NewErrorCode(ErrOperationFailed)
	ConflictError         = NewErrorCode(ErrConflict)

	// 认证错误
	UnauthorizedError      = NewErrorCode(ErrUnauthorized)
	TokenExpiredError      = NewErrorCode(ErrTokenExpired)
	TokenInvalidError      = NewErrorCode(ErrTokenInvalid)
	LoginFailedError       = NewErrorCode(ErrLoginFailed)
	PasswordIncorrectError = NewErrorCode(ErrPasswordIncorrect)
	AccountDisabledError   = NewErrorCode(ErrAccountDisabled)

	// 权限错误
	PermissionDeniedError = NewErrorCode(ErrPermissionDenied)
	TenantPermissionError = NewErrorCode(ErrTenantPermission)
	DataPermissionError   = NewErrorCode(ErrDataPermission)
	RoleNotExistError     = NewErrorCode(ErrRoleNotExist)
	MenuNotExistError     = NewErrorCode(ErrMenuNotExist)

	// 租户隔离错误
	TenantIDMissingError      = NewErrorCode(ErrTenantIDMissing)
	CrossTenantOperationError = NewErrorCode(ErrCrossTenantOperation)
	TenantNotExistError       = NewErrorCode(ErrTenantNotExist)
	TenantDisabledError       = NewErrorCode(ErrTenantDisabled)
	TenantQuotaExceededError  = NewErrorCode(ErrTenantQuotaExceeded)

	// 业务错误
	UserNotExistError       = NewErrorCode(ErrUserNotExist)
	UserAlreadyExistError   = NewErrorCode(ErrUserAlreadyExist)
	EmailAlreadyUsedError   = NewErrorCode(ErrEmailAlreadyUsed)
	PhoneAlreadyUsedError   = NewErrorCode(ErrPhoneAlreadyUsed)
	DepartmentNotExistError = NewErrorCode(ErrDepartmentNotExist)
	PositionNotExistError   = NewErrorCode(ErrPositionNotExist)

	// 应用市场错误
	PluginNotFoundError         = NewErrorCode(ErrPluginNotFound)
	PluginNotApprovedError      = NewErrorCode(ErrPluginNotApproved)
	PluginAlreadyInstalledError = NewErrorCode(ErrPluginAlreadyInstalled)
	PluginNotInstalledError     = NewErrorCode(ErrPluginNotInstalled)
	PluginInstallFailedError    = NewErrorCode(ErrPluginInstallFailed)
	PluginUninstallFailedError  = NewErrorCode(ErrPluginUninstallFailed)
	PluginVersionConflictError  = NewErrorCode(ErrPluginVersionConflict)
	PluginDisabledError         = NewErrorCode(ErrPluginDisabled)
	PluginVersionNotFoundError  = NewErrorCode(ErrPluginVersionNotFound)
	DeveloperNotFoundError      = NewErrorCode(ErrDeveloperNotFound)
	DeveloperPendingError       = NewErrorCode(ErrDeveloperPending)
	DeveloperSuspendedError     = NewErrorCode(ErrDeveloperSuspended)
	DeveloperAlreadyExistError  = NewErrorCode(ErrDeveloperAlreadyExist)
	CategoryNotFoundError       = NewErrorCode(ErrCategoryNotFound)
	CategoryAlreadyExistError   = NewErrorCode(ErrCategoryAlreadyExist)
	TagNotFoundError            = NewErrorCode(ErrTagNotFound)
	TagAlreadyExistError        = NewErrorCode(ErrTagAlreadyExist)
	SlugAlreadyExistError       = NewErrorCode(ErrSlugAlreadyExist)

	// 系统错误
	InternalServerError  = NewErrorCode(ErrInternalServer)
	DatabaseError        = NewErrorCode(ErrDatabase)
	CacheError           = NewErrorCode(ErrCache)
	FileSystemError      = NewErrorCode(ErrFileSystem)
	NetworkError         = NewErrorCode(ErrNetwork)
	ExternalServiceError = NewErrorCode(ErrExternalService)
)

type ServiceError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewServiceError(code int, message string) *ServiceError {
	if message == "" {
		message = GetErrorMessage(code)
	}
	return &ServiceError{
		Code:    code,
		Message: message,
	}
}

func (e *ServiceError) Error() string {
	return e.Message
}

func GetServiceErrorCode(err error) int {
	if se, ok := err.(*ServiceError); ok {
		return se.Code
	}
	return ErrInternalServer
}
