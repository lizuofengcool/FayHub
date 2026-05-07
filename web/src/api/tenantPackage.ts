import request from './request'

export interface TenantPackage {
  id: number
  name: string
  code: string
  status: number
  sort: number
  remark: string
  max_users: number
  max_storage_mb: number
  max_plugins: number
  created_at: string
  updated_at: string
}

export interface TenantPackageListParams {
  page?: number
  page_size?: number
  name?: string
  status?: number
}

export interface TenantPackageForm {
  name: string
  code: string
  status: number
  sort: number
  remark: string
  max_users: number
  max_storage_mb: number
  max_plugins: number
  menu_ids: number[]
}

export function getTenantPackageList(params: TenantPackageListParams) {
  return request.get('/tenant-packages', { params })
}

export function getAllTenantPackages() {
  return request.get('/tenant-packages/all')
}

export function getTenantPackage(id: number) {
  return request.get(`/tenant-packages/${id}`)
}

export function createTenantPackage(data: TenantPackageForm) {
  return request.post('/tenant-packages', data)
}

export function updateTenantPackage(id: number, data: TenantPackageForm) {
  return request.put(`/tenant-packages/${id}`, data)
}

export function deleteTenantPackage(id: number) {
  return request.delete(`/tenant-packages/${id}`)
}
