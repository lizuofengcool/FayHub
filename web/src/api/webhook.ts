import request, { type PageParams, type PageResult } from './request'

export interface WebhookSubscription {
  id: number
  name: string
  url: string
  secret: string
  events: string[]
  is_active: boolean
  description: string
  created_at: string
  updated_at: string
}

export interface WebhookDelivery {
  id: number
  subscription_id: number
  event: string
  payload: Record<string, any>
  status: string
  status_code: number
  response_body: string
  attempts: number
  next_retry_at: string
  delivered_at: string
  created_at: string
}

export interface WebhookStats {
  total_deliveries: number
  success_count: number
  failed_count: number
  pending_count: number
  success_rate: number
}

export interface CreateSubscriptionRequest {
  name: string
  url: string
  secret?: string
  events: string[]
  description?: string
}

export interface UpdateSubscriptionRequest {
  name?: string
  url?: string
  secret?: string
  events?: string[]
  is_active?: boolean
  description?: string
}

const webhookApi = {
  listSubscriptions(params?: PageParams) {
    return request.get<any, { code: number; data: PageResult<WebhookSubscription> }>('/webhooks/subscriptions', { params })
  },

  getSubscription(id: number) {
    return request.get<any, { code: number; data: WebhookSubscription }>(`/webhooks/subscriptions/${id}`)
  },

  createSubscription(data: CreateSubscriptionRequest) {
    return request.post('/webhooks/subscriptions', data)
  },

  updateSubscription(id: number, data: UpdateSubscriptionRequest) {
    return request.put(`/webhooks/subscriptions/${id}`, data)
  },

  deleteSubscription(id: number) {
    return request.delete(`/webhooks/subscriptions/${id}`)
  },

  listDeliveries(params?: PageParams & { subscription_id?: number; event?: string; status?: string }) {
    return request.get<any, { code: number; data: PageResult<WebhookDelivery> }>('/webhooks/deliveries', { params })
  },

  redeliver(deliveryId: number) {
    return request.post(`/webhooks/deliveries/${deliveryId}/redeliver`)
  },

  getDeliveryStats() {
    return request.get<any, { code: number; data: WebhookStats }>('/webhooks/deliveries/stats')
  }
}

export default webhookApi
