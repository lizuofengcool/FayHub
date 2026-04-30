import request, { type PageResult } from './request'

export interface AuditLog {
  id: number
  user_id: number
  username: string
  action: string
  resource: string
  resource_id: string
  detail: Record<string, any>
  ip: string
  user_agent: string
  request_id: string
  status_code: number
  success: boolean
  error_msg: string
  duration: number
  method: string
  path: string
  created_at: string
}

export interface AuditStats {
  total: number
  today: number
  success_rate: number
  top_actions: Array<{ action: string; count: number }>
  top_users: Array<{ username: string; count: number }>
}

export interface AuditLogParams {
  page?: number
  page_size?: number
  user_id?: number
  action?: string
  resource?: string
  resource_id?: string
  success?: boolean
  start_time?: string
  end_time?: string
  ip?: string
  path?: string
}

const auditApi = {
  listLogs(params?: AuditLogParams) {
    return request.get<any, { code: number; data: PageResult<AuditLog> }>('/audit/logs', { params })
  },

  getLog(id: number) {
    return request.get<any, { code: number; data: AuditLog }>(`/audit/logs/${id}`)
  },

  getStats() {
    return request.get<any, { code: number; data: AuditStats }>('/audit/stats')
  },

  cleanup(data: { before_time: string }) {
    return request.post('/audit/cleanup', data)
  }
}

export default auditApi
