import request, { type ApiResponse } from './request'

export interface LoginParams {
  username: string
  password: string
}

export interface RegisterParams {
  username: string
  password: string
  email?: string
  phone?: string
  real_name?: string
}

export interface LoginResponse {
  user_id: number
  username: string
  role: string
  tenant_id: number
  token: string
}

export interface RegisterResponse {
  user_id: number
  username: string
  role: string
  tenant_id: number
  token: string
}

export interface UserInfo {
  id: number
  user_id: number
  username: string
  role: string
  tenant_id: number
  avatar?: string
}

export interface CaptchaResponse {
  captcha_key: string
  captcha_code: string
  expires_in: number
}

const authApi = {
  login(params: LoginParams): Promise<ApiResponse<LoginResponse>> {
    return request.post('/auth/login', params)
  },

  register(params: RegisterParams): Promise<ApiResponse<RegisterResponse>> {
    return request.post('/auth/register', params)
  },

  getCaptcha(): Promise<ApiResponse<CaptchaResponse>> {
    return request.get('/auth/captcha')
  },

  logout(): Promise<ApiResponse<null>> {
    return request.post('/auth/logout')
  },

  refreshToken(token: string): Promise<ApiResponse<{ token: string }>> {
    return request.post('/auth/refresh', { token })
  },

  getCurrentUser(): Promise<ApiResponse<UserInfo>> {
    return request.get('/auth/me')
  }
}

export default authApi
