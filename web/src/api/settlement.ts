import request, { type PageResult, type PageParams } from './request'

export interface SettlementRecord {
  id: number
  tenant_id: number
  order_no: string
  order_id: number
  total_amount: number
  platform_amount: number
  tenant_amount: number
  platform_rate: number
  status: string
  settled_at: string | null
  settlement_no: string
  fail_reason: string
  created_at: string
  updated_at: string
}

export interface SettlementConfig {
  id: number
  tenant_id: number
  platform_rate: number
  min_amount: number
  status: number
  created_at: string
  updated_at: string
}

export interface SettlementStats {
  total_amount: number
  platform_amount: number
  tenant_amount: number
  pending_count: number
  settled_count: number
  failed_count: number
}

export interface CreateSettlementRequest {
  order_no: string
  total_amount: number
}

export interface UpdateSettlementConfigRequest {
  platform_rate: number
  min_amount: number
}

const settlementApi = {
  createSettlement(data: CreateSettlementRequest) {
    return request.post('/settlement', data)
  },

  getSettlementConfig() {
    return request.get<any, { code: number; data: SettlementConfig }>('/settlement/config')
  },

  updateSettlementConfig(data: UpdateSettlementConfigRequest) {
    return request.put('/settlement/config', data)
  },

  processSettlement(settlementNo: string) {
    return request.post(`/settlement/process/${settlementNo}`)
  },

  getSettlementStats() {
    return request.get<any, { code: number; data: SettlementStats }>('/settlement/stats')
  },

  listSettlements(params?: PageParams & { status?: string }) {
    return request.get<any, { code: number; data: PageResult<SettlementRecord> }>('/settlement/records', { params })
  }
}

export default settlementApi
