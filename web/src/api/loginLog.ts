import request, { type PageResult } from './request'

export interface LoginLog {
  id: number
  user_id: number
  username: string
  tenant_id: number
  login_status: string
  login_ip: string
  login_location: string
  browser: string
  os: string
  login_time: string
  logout_time: string | null
  msg: string
}

export interface LoginLogParams {
  page?: number
  page_size?: number
  username?: string
  login_status?: string
  login_ip?: string
  start_time?: string
  end_time?: string
}

const loginLogApi = {
  listLogs(params?: LoginLogParams) {
    return request.get<any, { code: number; data: PageResult<LoginLog> }>('/login-logs', { params })
  },

  cleanup(days: number) {
    return request.post('/login-logs/cleanup', null, { params: { days } })
  },
}

export default loginLogApi
