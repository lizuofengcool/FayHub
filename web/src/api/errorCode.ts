import request, { type PageResult } from './request'

export interface ErrorCodeItem {
  id: number
  code: number
  name: string
  msg: string
  status: number
  created_at: string
  updated_at: string
}

export interface ErrorCodeParams {
  page?: number
  page_size?: number
  name?: string
  code?: number
  status?: number
}

const errorCodeApi = {
  list(params?: ErrorCodeParams) {
    return request.get<any, { code: number; data: PageResult<ErrorCodeItem> }>('/error-codes', { params })
  },

  create(data: Partial<ErrorCodeItem>) {
    return request.post('/error-codes', data)
  },

  update(id: number, data: Partial<ErrorCodeItem>) {
    return request.put(`/error-codes/${id}`, data)
  },

  delete(id: number) {
    return request.delete(`/error-codes/${id}`)
  },

  refreshCache() {
    return request.post('/error-codes/refresh-cache')
  },
}

export default errorCodeApi
