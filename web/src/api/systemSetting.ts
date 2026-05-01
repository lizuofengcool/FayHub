import request from './request'

export interface DomainSettings {
  admin_url: string
  market_url: string
  dev_url: string
  api_url: string
  sso_url: string
  www_url: string
}

export interface PaymentGatewaySettings {
  notify_base_url: string
  order_expire_min: number
  wechat_gateway_url: string
  alipay_gateway_url: string
  alipay_sandbox_url: string
}

export interface SecuritySettings {
  max_login_attempts: number
  lock_duration_min: number
}

export interface ServerInfo {
  port: number
  mode: string
}

export interface SystemSettings {
  domains: DomainSettings
  payment: PaymentGatewaySettings
  security: SecuritySettings
  server: ServerInfo
}

const systemSettingApi = {
  getSettings() {
    return request.get<any, { code: number; data: SystemSettings }>('/system/settings')
  },

  updateSettings(data: Partial<SystemSettings>) {
    return request.put('/system/settings', data)
  }
}

export default systemSettingApi
