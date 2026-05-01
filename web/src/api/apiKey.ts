import request from './request'

export interface APIKey {
  id: number
  tenant_id: number
  user_id: number
  name: string
  key_prefix: string
  secret?: string
  permissions: string
  rate_limit: number
  expires_at: string | null
  last_used_at: string | null
  status: number
  created_at: string
  updated_at: string
}

export interface APIKeyPermission {
  resource: string
  action: string
}

export interface CreateAPIKeyRequest {
  name: string
  permissions?: APIKeyPermission[]
  rate_limit?: number
  expires_at?: string
}

const apiKeyApi = {
  createAPIKey(data: CreateAPIKeyRequest) {
    return request.post<any, { code: number; data: APIKey }>('/api-keys', data)
  },

  listAPIKeys() {
    return request.get<any, { code: number; data: APIKey[] }>('/api-keys')
  },

  deleteAPIKey(id: number) {
    return request.delete(`/api-keys/${id}`)
  }
}

export default apiKeyApi
