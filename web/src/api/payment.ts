import request, { type PageParams, type PageResult } from './request'

export interface WechatConfig {
  enabled: boolean
  mch_id: string
  app_id: string
  api_key: string
  serial_no: string
  notify_url: string
}

export interface AlipayConfig {
  enabled: boolean
  app_id: string
  private_key: string
  public_key: string
  notify_url: string
  sandbox: boolean
}

export interface PaymentConfig {
  wechat: WechatConfig
  alipay: AlipayConfig
}

export interface PaymentOrder {
  order_no: string
  plugin_name: string
  buyer: string
  amount: string
  status: string
  pay_method: string
  created_at: string
}

export interface PaymentStats {
  totalAmount: string
  totalCount: number
  platformIncome: string
  pendingSettlement: string
}

export interface TransactionListParams extends PageParams {
  status?: string
  start_date?: string
  end_date?: string
}

const paymentApi = {
  getConfig() {
    return request.get<any, { code: number; data: PaymentConfig }>('/payment/config')
  },

  updateConfig(data: Partial<PaymentConfig>) {
    return request.put('/payment/config', data)
  },

  createOrder(data: { plugin_id: string; amount: number; pay_method: string }) {
    return request.post('/payment/orders', data)
  },

  listTransactions(params?: TransactionListParams) {
    return request.get<any, { code: number; data: PageResult<PaymentOrder> & { stats?: PaymentStats } }>('/payment/transactions', { params })
  },

  getStats() {
    return request.get<any, { code: number; data: PaymentStats }>('/payment/transactions/stats')
  },

  refund(data: { order_no: string; reason: string }) {
    return request.post('/payment/refund', data)
  }
}

export default paymentApi
