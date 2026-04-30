import request from './request'

export interface SSOTokenExchangeRequest {
  code: string
}

export interface SSOTokenResponse {
  access_token: string
  token_type: string
  expires_in: number
  user: {
    id: number
    username: string
    tenant_id: number
  }
}

const ssoApi = {
  // 获取授权码
  getAuthorizationCode() {
    return request.get('/sso/authorize')
  },

  // 授权码换令牌（供市场调用）
  exchangeToken(code: string) {
    return request.post('/sso/token', { code })
  },

  // 验证令牌（供市场调用）
  verifyToken(token: string) {
    return request.post('/sso/verify', { token })
  }
}

export default ssoApi
