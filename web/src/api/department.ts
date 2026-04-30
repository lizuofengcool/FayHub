import request, { type ApiResponse } from './request'

export interface Department {
  id: number
  name: string
  parent_id: number
  sort: number
  status: number
  leader_id: number
  children?: Department[]
  created_at: string
  updated_at: string
}

export interface CreateDeptParams {
  name: string
  parent_id?: number
  sort?: number
  leader_id?: number
}

export interface UpdateDeptParams {
  name?: string
  parent_id?: number
  sort?: number
  status?: number
  leader_id?: number
}

const deptApi = {
  getTree(): Promise<ApiResponse<Department[]>> {
    return request.get('/departments/tree')
  },

  create(params: CreateDeptParams): Promise<ApiResponse<Department>> {
    return request.post('/departments', params)
  },

  update(id: number, params: UpdateDeptParams): Promise<ApiResponse<Department>> {
    return request.put(`/departments/${id}`, params)
  },

  delete(id: number): Promise<ApiResponse<null>> {
    return request.delete(`/departments/${id}`)
  },

  assignUser(userId: number, deptId: number): Promise<ApiResponse<null>> {
    return request.post(`/departments/${deptId}/users/${userId}`)
  },

  removeUser(userId: number, deptId: number): Promise<ApiResponse<null>> {
    return request.delete(`/departments/${deptId}/users/${userId}`)
  }
}

export default deptApi
