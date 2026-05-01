import request from './request'

export interface BackupRecord {
  id: number
  filename: string
  file_size: number
  status: string
  created_at: string
}

const backupApi = {
  createBackup() {
    return request.post<any, { code: number; data: BackupRecord }>('/backups')
  },

  listBackups() {
    return request.get<any, { code: number; data: { list: BackupRecord[]; total: number } }>('/backups')
  },

  deleteBackup(id: number) {
    return request.delete(`/backups/${id}`)
  },

  downloadBackup(id: number) {
    return request.get(`/backups/${id}/download`, { responseType: 'blob' })
  },

  restoreBackup(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    return request.post('/backups/restore', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  }
}

export default backupApi
