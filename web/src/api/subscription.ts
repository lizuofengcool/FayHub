import request, { type ApiResponse } from './request'

export interface Subscription {
  id: number
  tenant_id: number
  package_id: number
  package_name: string
  status: string
  start_date: string
  end_date: string
  trial_end_date: string | null
  auto_renew: number
  max_users: number
  max_storage: number
  current_users: number
  current_storage: number
  created_at: string
  updated_at: string
}

export interface SubscriptionInvoice {
  id: number
  subscription_id: number
  tenant_id: number
  amount: number
  currency: string
  status: string
  billing_period: string
  paid_at: string | null
  due_date: string
  created_at: string
  updated_at: string
}

export interface SubscriptionListResult {
  list: Subscription[]
  total: number
  page: number
  page_size: number
}

export interface InvoiceListResult {
  list: SubscriptionInvoice[]
  total: number
  page: number
  page_size: number
}

const subscriptionApi = {
  list(page = 1, pageSize = 20): Promise<ApiResponse<SubscriptionListResult>> {
    return request.get('/subscriptions', { params: { page, page_size: pageSize } })
  },

  getById(id: number): Promise<ApiResponse<Subscription>> {
    return request.get(`/subscriptions/${id}`)
  },

  create(data: Partial<Subscription>): Promise<ApiResponse<Subscription>> {
    return request.post('/subscriptions', data)
  },

  update(id: number, data: Partial<Subscription>): Promise<ApiResponse<Subscription>> {
    return request.put(`/subscriptions/${id}`, data)
  },

  delete(id: number): Promise<ApiResponse<null>> {
    return request.delete(`/subscriptions/${id}`)
  },

  cancel(id: number): Promise<ApiResponse<null>> {
    return request.post(`/subscriptions/${id}/cancel`)
  },

  renew(id: number, months: number): Promise<ApiResponse<null>> {
    return request.post(`/subscriptions/${id}/renew`, { months })
  },

  getInvoices(id: number, page = 1, pageSize = 20): Promise<ApiResponse<InvoiceListResult>> {
    return request.get(`/subscriptions/${id}/invoices`, { params: { page, page_size: pageSize } })
  }
}

export default subscriptionApi
