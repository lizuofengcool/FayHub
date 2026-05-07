import request, { type ApiResponse } from './request'

export interface CronJob {
  id: number
  tenant_id: number
  name: string
  cron_expr: string
  command: string
  description: string
  status: number
  last_run_at: string | null
  next_run_at: string | null
  created_at: string
  updated_at: string
}

export interface CronJobLog {
  id: number
  job_id: number
  tenant_id: number
  status: string
  output: string
  error: string
  started_at: string
  ended_at: string | null
  created_at: string
}

export interface CronJobListResult {
  list: CronJob[]
  total: number
  page: number
  page_size: number
}

export interface CronJobLogListResult {
  list: CronJobLog[]
  total: number
  page: number
  page_size: number
}

const cronJobApi = {
  list(page = 1, pageSize = 20): Promise<ApiResponse<CronJobListResult>> {
    return request.get('/cron-jobs', { params: { page, page_size: pageSize } })
  },

  getById(id: number): Promise<ApiResponse<CronJob>> {
    return request.get(`/cron-jobs/${id}`)
  },

  create(data: Partial<CronJob>): Promise<ApiResponse<CronJob>> {
    return request.post('/cron-jobs', data)
  },

  update(id: number, data: Partial<CronJob>): Promise<ApiResponse<CronJob>> {
    return request.put(`/cron-jobs/${id}`, data)
  },

  delete(id: number): Promise<ApiResponse<null>> {
    return request.delete(`/cron-jobs/${id}`)
  },

  toggleStatus(id: number): Promise<ApiResponse<null>> {
    return request.put(`/cron-jobs/${id}/toggle`)
  },

  executeOnce(id: number): Promise<ApiResponse<null>> {
    return request.post(`/cron-jobs/${id}/execute`)
  },

  getLogs(id: number, page = 1, pageSize = 20): Promise<ApiResponse<CronJobLogListResult>> {
    return request.get(`/cron-jobs/${id}/logs`, { params: { page, page_size: pageSize } })
  }
}

export default cronJobApi
