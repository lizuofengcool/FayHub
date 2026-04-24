import axios from 'axios'

const service = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

service.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

service.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.code !== 0 && res.code !== 200) {
      return Promise.reject(new Error(res.msg || '请求失败'))
    }
    return res
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/'
    }
    return Promise.reject(error)
  }
)

export interface LoginParams {
  username: string
  password: string
  captcha?: string
}

export interface RegisterParams {
  username: string
  password: string
  email: string
  phone: string
  real_name: string
  tenant_id: number
}

export interface UserInfo {
  id?: number
  user_id?: number
  username: string
  nickname?: string
  role: string
  tenant_id?: number
}

export interface BackendLoginResponse {
  user_id: number
  username: string
  role: string
  token: string
}

export interface ApiResponse<T = any> {
  code: number
  data: T
  msg: string
}

const authApi = {
  login(params: LoginParams): Promise<ApiResponse<BackendLoginResponse>> {
    return service.post('/auth/login', params)
  },

  register(params: RegisterParams): Promise<ApiResponse<BackendLoginResponse>> {
    return service.post('/auth/register', params)
  },

  logout(): Promise<ApiResponse<null>> {
    return service.post('/auth/logout')
  },

  refreshToken(token: string): Promise<ApiResponse<{ token: string }>> {
    return service.post('/auth/refresh', { token })
  },

  getCurrentUser(): Promise<ApiResponse<UserInfo>> {
    return service.get('/auth/me')
  }
}

export default authApi
