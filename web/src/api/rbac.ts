import request, { type ApiResponse, type PageResult, type PageParams } from './request'

export interface Role {
  id: number
  name: string
  type: number
  description: string
  status: number
  data_scope: number
  tenant_id: number
  created_at: string
  updated_at: string
}

export interface CreateRoleParams {
  name: string
  type: number
  description?: string
  data_scope?: number
}

export interface UpdateRoleParams {
  name?: string
  type?: number
  description?: string
  data_scope?: number
}

export interface RoleListParams extends PageParams {}

export interface AssignRoleParams {
  user_id: number
  role_id: number
}

export interface RemoveRoleParams {
  user_id: number
  role_id: number
}

export interface Permission {
  menus: string[]
  apis: string[]
}

const rbacApi = {
  createRole(params: CreateRoleParams): Promise<ApiResponse<Role>> {
    return request.post('/rbac/roles', params)
  },

  getRoleList(params?: RoleListParams): Promise<ApiResponse<PageResult<Role>>> {
    return request.get('/rbac/roles', { params })
  },

  getRoleByID(roleID: number): Promise<ApiResponse<Role>> {
    return request.get(`/rbac/roles/${roleID}`)
  },

  updateRole(roleID: number, params: UpdateRoleParams): Promise<ApiResponse<Role>> {
    return request.put(`/rbac/roles/${roleID}`, params)
  },

  deleteRole(roleID: number): Promise<ApiResponse<null>> {
    return request.delete(`/rbac/roles/${roleID}`)
  },

  assignRoleToUser(params: AssignRoleParams): Promise<ApiResponse<null>> {
    return request.post('/rbac/assign-role', params)
  },

  removeRoleFromUser(params: RemoveRoleParams): Promise<ApiResponse<null>> {
    return request.post('/rbac/remove-role', params)
  },

  getUserRoles(userID: number): Promise<ApiResponse<Role[]>> {
    return request.get(`/rbac/users/${userID}/roles`)
  },

  getUserPermissions(userID: number): Promise<ApiResponse<Permission>> {
    return request.get(`/rbac/users/${userID}/permissions`)
  }
}

export default rbacApi
