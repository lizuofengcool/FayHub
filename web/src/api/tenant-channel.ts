import request, { type PageResult } from './request'

export interface TenantChannelConfig {
  id: number
  tenant_id?: number
  channel_type: string
  channel_name?: string
  app_id?: string
  app_secret?: string
  merchant_id?: string
  pay_public_key?: string
  pay_private_key?: string
  cert_serial_no?: string
  token?: string
  encoding_aes_key?: string
  extra?: string
  status: number
  created_at?: string
  updated_at?: string
}

export interface UserThirdParty {
  id: number
  user_id: number
  tenant_id: number
  channel_type: string
  open_id: string
  union_id?: string
  nickname?: string
  avatar?: string
  gender?: number
  country?: string
  province?: string
  city?: string
  bind_at?: number
  last_active_at?: number
  extra?: string
  created_at?: string
  updated_at?: string
}

export interface ChannelConfigParams {
  page?: number
  page_size?: number
  channel_type?: string
  status?: number
}

export interface UserBindingParams {
  user_id?: number
  channel_type?: string
}

const channelApi = {
  listConfigs(params?: ChannelConfigParams) {
    return request.get<any, { code: number; data: PageResult<TenantChannelConfig> }>('/tenant-channel/configs', { params })
  },

  getConfig(id: number) {
    return request.get<any, { code: number; data: TenantChannelConfig }>(`/tenant-channel/configs/${id}`)
  },

  getConfigByType(channelType: string) {
    return request.get<any, { code: number; data: TenantChannelConfig }>(`/tenant-channel/configs/type/${channelType}`)
  },

  createConfig(data: Partial<TenantChannelConfig>) {
    return request.post('/tenant-channel/configs', data)
  },

  updateConfig(id: number, data: Partial<TenantChannelConfig>) {
    return request.put(`/tenant-channel/configs/${id}`, data)
  },

  deleteConfig(id: number) {
    return request.delete(`/tenant-channel/configs/${id}`)
  },

  getUserBindings(params?: UserBindingParams) {
    return request.get<any, { code: number; data: { list: UserThirdParty[] } }>('/tenant-channel/bindings', { params })
  },

  deleteUserBinding(id: number) {
    return request.delete(`/tenant-channel/bindings/${id}`)
  },
}

export default channelApi
