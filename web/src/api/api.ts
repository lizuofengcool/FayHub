import request, { type ApiResponse, type PageResult, type PageParams } from './request'

export interface ApiItem {
  id: number
  path: string
  method: string
  description: string
  group: string
  status: number
  tenant_id: number
  created_at: string
  updated_at: string
}

export interface CreateApiParams {
  path: string
  method: string
  description?: string
  group?: string
  status?: number
}

export interface UpdateApiParams {
  path?: string
  method?: string
  description?: string
  group?: string
  status?: number
}

export interface ApiListParams extends PageParams {
  group?: string
  method?: string
}

export interface AssignRoleApisParams {
  role_id: number
  api_ids: number[]
}

const apiApi = {
  createApi(params: CreateApiParams): Promise<ApiResponse<ApiItem>> {
    return request.post('/apis', params)
  },

  getApiList(params?: ApiListParams): Promise<ApiResponse<PageResult<ApiItem>>> {
    return request.get('/apis', { params })
  },

  getApiByID(apiID: number): Promise<ApiResponse<ApiItem>> {
    return request.get(`/apis/${apiID}`)
  },

  updateApi(apiID: number, params: UpdateApiParams): Promise<ApiResponse<ApiItem>> {
    return request.put(`/apis/${apiID}`, params)
  },

  deleteApi(apiID: number): Promise<ApiResponse<null>> {
    return request.delete(`/apis/${apiID}`)
  },

  assignRoleApis(params: AssignRoleApisParams): Promise<ApiResponse<null>> {
    return request.post('/apis/assign-role', params)
  },

  getRoleApis(roleID: number): Promise<ApiResponse<ApiItem[]>> {
    return request.get(`/apis/role/${roleID}`)
  }
}

export default apiApi
