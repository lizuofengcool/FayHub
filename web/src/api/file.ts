import request, { type PageResult, type PageParams } from './request'

export interface FileRecord {
  id: number
  file_name: string
  original_name: string
  file_key: string
  file_size: number
  mime_type: string
  storage_driver: string
  url: string
  user_id: number
  created_at: string
  deleted_at?: string
}

export interface FileListParams extends PageParams {
  mime_type?: string
}

export interface UploadResult {
  id: number
  file_name: string
  original_name: string
  file_key: string
  file_size: number
  mime_type: string
  url: string
}

const fileApi = {
  upload(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    return request.post<any, { code: number; data: UploadResult }>('/files/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  list(params?: FileListParams) {
    return request.get<any, { code: number; data: PageResult<FileRecord> }>('/files/list', { params })
  },

  download(id: number) {
    return request.get(`/files/${id}`, { responseType: 'blob' })
  },

  delete(id: number) {
    return request.delete(`/files/${id}`)
  }
}

export default fileApi
