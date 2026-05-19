import axios from 'axios'
import { ElMessage } from 'element-plus'

let isRefreshing = false
let refreshSubscribers: Array<(token: string) => void> = []

function onTokenRefreshed(newToken: string) {
  refreshSubscribers.forEach((cb) => cb(newToken))
  refreshSubscribers = []
}

function addRefreshSubscriber(cb: (token: string) => void) {
  refreshSubscribers.push(cb)
}

function getToken(): string {
  return localStorage.getItem('fayhub_token') || ''
}

function getRefreshToken(): string {
  return localStorage.getItem('fayhub_refresh_token') || ''
}

function isTokenExpiringSoon(token: string): boolean {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    const exp = payload.exp * 1000
    const threshold = 5 * 60 * 1000
    return exp - Date.now() < threshold
  } catch {
    return true
  }
}

const service = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  },
  withCredentials: true
})

service.interceptors.request.use(
  async (config) => {
    const token = getToken()
    if (token) {
      if (isTokenExpiringSoon(token) && !config.url?.includes('/auth/refresh')) {
        if (!isRefreshing) {
          isRefreshing = true
          try {
            const refreshToken = getRefreshToken()
            const currentToken = getToken()
            const tokenToSend = refreshToken || currentToken
            if (!tokenToSend) {
              throw new Error('No token available for refresh')
            }
            const res = await axios.post('/api/auth/refresh', { token: tokenToSend }, { withCredentials: true })
            if (res.data?.code === 200 && res.data?.data?.token) {
              const newToken = res.data.data.token
              const newRefreshToken = res.data.data.refresh_token
              localStorage.setItem('fayhub_token', newToken)
              if (newRefreshToken) {
                localStorage.setItem('fayhub_refresh_token', newRefreshToken)
              }
              onTokenRefreshed(newToken)
            }
          } catch {
            localStorage.removeItem('fayhub_token')
            localStorage.removeItem('fayhub_refresh_token')
            localStorage.removeItem('userInfo')
            window.location.href = '/'
            return Promise.reject(new Error('Token刷新失败，请重新登录'))
          } finally {
            isRefreshing = false
          }
        }
        if (isRefreshing) {
          return new Promise((resolve) => {
            addRefreshSubscriber(() => {
              const newToken = getToken()
              if (newToken) {
                config.headers['Authorization'] = `Bearer ${newToken}`
              }
              resolve(config)
            })
          })
        }
      }
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
    if (res.code !== 200) {
      if (res.code === 41001 || res.code === 41002) {
        localStorage.removeItem('fayhub_token')
        localStorage.removeItem('fayhub_refresh_token')
        localStorage.removeItem('userInfo')
        window.location.href = '/'
      }
      return Promise.reject(new Error(res.msg || '请求失败'))
    }
    return res
  },
  (error) => {
    const res = error.response?.data
    if (res && (res.code === 41001 || res.code === 41002)) {
      localStorage.removeItem('fayhub_token')
      localStorage.removeItem('fayhub_refresh_token')
      localStorage.removeItem('userInfo')
      window.location.href = '/'
    } else if (error.response?.status === 401) {
      const isLoginRequest = error.config?.url?.includes('/auth/login')
      if (!isLoginRequest) {
        localStorage.removeItem('fayhub_token')
        localStorage.removeItem('fayhub_refresh_token')
        localStorage.removeItem('userInfo')
        window.location.href = '/'
      }
    } else {
      ElMessage.error(res?.msg || error.message || '网络请求失败')
    }
    return Promise.reject(error)
  }
)

export interface ApiResponse<T = unknown> {
  code: number
  data: T
  msg: string
}

export interface PageResult<T = unknown> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export interface PageParams {
  page?: number
  page_size?: number
  keyword?: string
}

export default service
