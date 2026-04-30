import request, { type ApiResponse, type PageResult, type PageParams } from './request'

export interface User {
  id: number
  username: string
  real_name: string
  email: string
  phone: string
  avatar: string
  status: number
  role: string
  tenant_id: number
  created_at: string
  updated_at: string
}

export interface CreateUserParams {
  username: string
  password: string
  real_name?: string
  email?: string
  phone?: string
  status?: number
}

export interface UpdateUserParams {
  real_name?: string
  email?: string
  phone?: string
  status?: number
}

export interface UserListParams extends PageParams {
  status?: number
}

export interface ChangePasswordParams {
  old_password: string
  new_password: string
}

export interface ResetPasswordParams {
  new_password: string
}

const userApi = {
  createUser(params: CreateUserParams): Promise<ApiResponse<User>> {
    return request.post('/users', params)
  },

  getUserList(params?: UserListParams): Promise<ApiResponse<PageResult<User>>> {
    return request.get('/users', { params })
  },

  getUser(id: number): Promise<ApiResponse<User>> {
    return request.get(`/users/${id}`)
  },

  getProfile(): Promise<ApiResponse<User>> {
    return request.get('/users/profile')
  },

  updateUser(id: number, params: UpdateUserParams): Promise<ApiResponse<User>> {
    return request.put(`/users/${id}`, params)
  },

  deleteUser(id: number): Promise<ApiResponse<null>> {
    return request.delete(`/users/${id}`)
  },

  changePassword(params: ChangePasswordParams): Promise<ApiResponse<null>> {
    return request.put('/users/change-password', params)
  },

  resetPassword(id: number, params: ResetPasswordParams): Promise<ApiResponse<null>> {
    return request.put(`/users/${id}/reset-password`, params)
  }
}

export default userApi
