import request, { type ApiResponse, type PageResult, type PageParams } from './request'

export interface Tenant {
  id: number
  name: string
  domain: string
  description: string
  status: number
  expired_at: number
  created_at: string
  updated_at: string
}

export interface CreateTenantParams {
  name: string
  domain?: string
  description?: string
  status?: number
  expired_at?: number
}

export interface UpdateTenantParams {
  name?: string
  domain?: string
  description?: string
  status?: number
  expired_at?: number
}

export interface TenantQuota {
  tenant_id: number
  max_users: number
  max_storage_mb: number
  max_plugins: number
  max_api_per_day: number
  used_users: number
  used_storage_mb: number
  used_plugins: number
  used_api_per_day: number
}

export interface QuotaCheckResult {
  allowed: boolean
  reason: string
  used: number
  max: number
}

export interface TenantListParams extends PageParams {
  status?: number
}

const tenantApi = {
  createTenant(params: CreateTenantParams): Promise<ApiResponse<Tenant>> {
    return request.post('/tenants', params)
  },

  getTenantList(params?: TenantListParams): Promise<ApiResponse<PageResult<Tenant>>> {
    return request.get('/tenants', { params })
  },

  getTenant(id: number): Promise<ApiResponse<Tenant>> {
    return request.get(`/tenants/${id}`)
  },

  updateTenant(id: number, params: UpdateTenantParams): Promise<ApiResponse<Tenant>> {
    return request.put(`/tenants/${id}`, params)
  },

  deleteTenant(id: number): Promise<ApiResponse<null>> {
    return request.delete(`/tenants/${id}`)
  },

  getTenantQuota(id: number): Promise<ApiResponse<TenantQuota>> {
    return request.get(`/tenants/${id}/quota`)
  },

  updateTenantQuota(id: number, data: Partial<TenantQuota>): Promise<ApiResponse<TenantQuota>> {
    return request.put(`/tenants/${id}/quota`, data)
  },

  syncTenantUsage(id: number): Promise<ApiResponse<TenantQuota>> {
    return request.post(`/tenants/${id}/quota/sync`)
  }
}

export default tenantApi
