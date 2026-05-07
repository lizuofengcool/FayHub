package model

// 渠道类型常量
const (
	ChannelTypeWechatMp       = "wechat_mp"       // 微信公众号
	ChannelTypeWechatMini     = "wechat_mini"     // 微信小程序
	ChannelTypeWechatPay      = "wechat_pay"      // 微信支付
	ChannelTypeAlipayMini     = "alipay_mini"     // 支付宝小程序
	ChannelTypeAlipayOfficial = "alipay_official" // 支付宝生活号
	ChannelTypeDouyinMini     = "douyin_mini"     // 抖音小程序
	ChannelTypeToutiaoMini    = "toutiao_mini"    // 头条小程序
)

// TenantChannelConfig 租户渠道配置表
// 用于存储租户的各平台渠道配置（微信公众号、小程序、支付等）
// 所有插件共享此配置，基座统一管理
type TenantChannelConfig struct {
	SnowflakeTenantModel
	ChannelType    string `gorm:"size:50;index:idx_tenant_channel;not null" json:"channel_type"` // 渠道类型：wechat_mp, wechat_mini, wechat_pay, alipay_mini...
	ChannelName    string `gorm:"size:100" json:"channel_name"`                                  // 渠道名称（用于展示，如"我的公众号"）
	AppID          string `gorm:"size:200" json:"app_id"`                                        // AppID / AppKey
	AppSecret      string `gorm:"size:500" json:"-"`                                             // AppSecret / AppSecret（加密存储，不返回给前端）
	MerchantID     string `gorm:"size:100" json:"merchant_id"`                                   // 商户号（支付渠道专用）
	PayPublicKey   string `gorm:"type:text" json:"-"`                                            // 支付公钥（加密存储，不返回给前端）
	PayPrivateKey  string `gorm:"type:text" json:"-"`                                            // 支付私钥（加密存储，不返回给前端）
	CertSerialNo   string `gorm:"size:100" json:"cert_serial_no"`                                // 证书序列号
	Token          string `gorm:"size:200" json:"-"`                                             // 服务器配置 Token（不返回给前端）
	EncodingAESKey string `gorm:"size:200" json:"-"`                                             // 消息加密密钥（不返回给前端）
	Extra          string `gorm:"type:text" json:"extra"`                                        // 扩展配置（JSON格式，存储其他渠道特有的字段）
	Status         int    `gorm:"default:1" json:"status"`                                       // 状态：1-启用，0-禁用
}

// UserThirdParty 用户第三方身份绑定表
// 用于关联 FayHub 底座用户与各平台的 OpenID/UnionID
// 基座统一管理，所有插件共享
type UserThirdParty struct {
	SnowflakeTenantModel
	UserID       int64  `gorm:"index:idx_user_channel;not null" json:"user_id"`              // FayHub 底座用户 ID
	ChannelType  string `gorm:"size:50;index:idx_user_channel;not null" json:"channel_type"` // 渠道类型：wechat_mp, wechat_mini...
	OpenID       string `gorm:"size:200;index:idx_openid;not null" json:"open_id"`           // 第三方平台 OpenID
	UnionID      string `gorm:"size:200;index:idx_unionid" json:"union_id"`                  // 第三方平台 UnionID（用于多账号打通）
	Nickname     string `gorm:"size:200" json:"nickname"`                                    // 昵称
	Avatar       string `gorm:"size:500" json:"avatar"`                                      // 头像
	Gender       int    `gorm:"default:0" json:"gender"`                                     // 性别：0-未知，1-男，2-女
	Country      string `gorm:"size:100" json:"country"`                                     // 国家
	Province     string `gorm:"size:100" json:"province"`                                    // 省份
	City         string `gorm:"size:100" json:"city"`                                        // 城市
	BindAt       int64  `gorm:"default:0" json:"bind_at"`                                    // 绑定时间戳
	LastActiveAt int64  `gorm:"default:0" json:"last_active_at"`                             // 最后活跃时间戳
	Extra        string `gorm:"type:text" json:"extra"`                                      // 扩展字段（JSON格式）
}

// TableName 设置表名
func (TenantChannelConfig) TableName() string {
	return "tenant_channel_configs"
}

// TableName 设置表名
func (UserThirdParty) TableName() string {
	return "user_third_parties"
}
