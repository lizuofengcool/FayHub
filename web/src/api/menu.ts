import request, { type ApiResponse, type PageResult, type PageParams } from './request'

export interface Menu {
  id: number
  parent_id: number
  title: string
  path: string
  component: string
  icon: string
  sort: number
  type: number
  status: number
  permission: string
  layout: string
  tenant_id: number
  created_at: string
  updated_at: string
  children?: Menu[]
}

export interface CreateMenuParams {
  parent_id?: number
  title: string
  path: string
  component?: string
  icon?: string
  sort?: number
  type?: number
  status?: number
  permission?: string
  layout?: 'embedded' | 'fullscreen'
}

export interface UpdateMenuParams {
  parent_id?: number
  title?: string
  path?: string
  component?: string
  icon?: string
  sort?: number
  type?: number
  status?: number
  permission?: string
  layout?: 'embedded' | 'fullscreen'
}

export interface MenuListParams extends PageParams {}

export interface AssignRoleMenusParams {
  role_id: number
  menu_ids: number[]
}

const menuApi = {
  createMenu(params: CreateMenuParams): Promise<ApiResponse<Menu>> {
    return request.post('/menus', params)
  },

  getMenuList(params?: MenuListParams): Promise<ApiResponse<PageResult<Menu>>> {
    return request.get('/menus', { params })
  },

  getMenuTree(): Promise<ApiResponse<Menu[]>> {
    return request.get('/menus/tree')
  },

  getMenuByID(menuID: number): Promise<ApiResponse<Menu>> {
    return request.get(`/menus/${menuID}`)
  },

  updateMenu(menuID: number, params: UpdateMenuParams): Promise<ApiResponse<Menu>> {
    return request.put(`/menus/${menuID}`, params)
  },

  deleteMenu(menuID: number): Promise<ApiResponse<null>> {
    return request.delete(`/menus/${menuID}`)
  },

  assignRoleMenus(params: AssignRoleMenusParams): Promise<ApiResponse<null>> {
    return request.post('/menus/assign-role', params)
  },

  getRoleMenus(roleID: number): Promise<ApiResponse<Menu[]>> {
    return request.get(`/menus/role/${roleID}`)
  }
}

export default menuApi
