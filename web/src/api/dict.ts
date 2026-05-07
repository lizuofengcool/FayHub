import request, { type PageResult } from './request'

export interface DictType {
  id: number
  dict_name: string
  dict_type: string
  status: number
  remark: string
  created_at: string
  updated_at: string
}

export interface DictData {
  id: number
  dict_type: string
  dict_label: string
  dict_value: string
  css_class: string
  list_class: string
  is_default: number
  sort: number
  status: number
  remark: string
  created_at: string
  updated_at: string
}

export interface DictTypeParams {
  page?: number
  page_size?: number
  dict_name?: string
  dict_type?: string
  status?: number
}

export interface DictDataParams {
  page?: number
  page_size?: number
  dict_type?: string
  dict_label?: string
  status?: number
}

const dictApi = {
  listTypes(params?: DictTypeParams) {
    return request.get<any, { code: number; data: PageResult<DictType> }>('/dict/types', { params })
  },

  getType(id: number) {
    return request.get<any, { code: number; data: DictType }>(`/dict/types/${id}`)
  },

  createType(data: Partial<DictType>) {
    return request.post('/dict/types', data)
  },

  updateType(id: number, data: Partial<DictType>) {
    return request.put(`/dict/types/${id}`, data)
  },

  deleteType(id: number) {
    return request.delete(`/dict/types/${id}`)
  },

  listData(params?: DictDataParams) {
    return request.get<any, { code: number; data: PageResult<DictData> }>('/dict/data', { params })
  },

  getDataByType(dictType: string) {
    return request.get<any, { code: number; data: DictData[] }>(`/dict/data/${dictType}`)
  },

  createData(data: Partial<DictData>) {
    return request.post('/dict/data', data)
  },

  updateData(id: number, data: Partial<DictData>) {
    return request.put(`/dict/data/${id}`, data)
  },

  deleteData(id: number) {
    return request.delete(`/dict/data/${id}`)
  },
}

export default dictApi
