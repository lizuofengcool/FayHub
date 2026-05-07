import request from './request'

export interface BackupRecord {
  id: number
  filename: string
  file_size: number
  status: string
  notes: string
  volumes: number
  created_at: string
}

export interface TableInfo {
  name: string
  comment: string
  row_count: number
  total_size: string
  index_size: string
  update_time: string
}

export interface ProcessInfo {
  pid: number
  user: string
  host: string
  database: string
  command: string
  time: string
  state: string
  query: string
}

export interface FieldVerifyResult {
  table_name: string
  field_count: number
  row_count: number
  status: string
  issues: string[]
}

export interface FieldInfo {
  name: string
  type: string
  nullable: string
  default: string
  comment: string
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
  },

  restoreBackupByID(id: number) {
    return request.post(`/backups/restore/${id}`)
  },

  listTables() {
    return request.get<any, { code: number; data: TableInfo[] }>('/backups/tables')
  },

  createBackupForTables(tables: string[]) {
    return request.post('/backups/tables', { tables })
  },

  executeSQL(sql: string, showErrors = false) {
    return request.post('/backups/execute', { sql, show_errors: showErrors })
  },

  executeWriteSQL(sql: string) {
    return request.post('/backups/sql/write', { sql })
  },

  listProcesses() {
    return request.get<any, { code: number; data: ProcessInfo[] }>('/backups/processes')
  },

  killProcess(pid: number) {
    return request.delete(`/backups/processes/${pid}`)
  },

  verifyFields() {
    return request.get<any, { code: number; data: FieldVerifyResult[] }>('/backups/verify')
  },

  replaceData(table: string, find: string, replace: string) {
    return request.post('/backups/replace', { table, find, replace })
  },

  advancedReplace(params: {
    table: string
    field?: string
    find: string
    replace: string
    replace_type: number
    condition?: string
    batch_size?: number
  }) {
    return request.post('/backups/replace/advanced', params)
  },

  dataTransfer(params: {
    source_table: string
    target_table: string
    condition?: string
    delete_source?: boolean
  }) {
    return request.post('/backups/transfer', params)
  },

  exportTable(table: string, format = 'csv') {
    return request.get('/backups/export', {
      params: { table, format },
      responseType: 'blob'
    })
  },

  advancedExport(params: {
    table: string
    fields?: string
    condition?: string
    time_field?: string
    from_date?: string
    to_date?: string
    order?: string
    format: string
    page_size?: number
    page?: number
  }) {
    return request.get('/backups/export/advanced', {
      params,
      responseType: 'blob'
    })
  },

  importData(table: string, file: File) {
    const formData = new FormData()
    formData.append('table', table)
    formData.append('file', file)
    return request.post('/backups/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  getTableFields(tableName: string) {
    return request.get<any, { code: number; data: FieldInfo[] }>(`/backups/tables/${tableName}/fields`)
  },

  getTableCount(tableName: string, condition?: string) {
    return request.get<any, { code: number; data: { total: number } }>(`/backups/tables/${tableName}/count`, {
      params: condition ? { condition } : {}
    })
  },

  previewTable(tableName: string, limit = 20) {
    return request.get<any, { code: number; data: { rows: any[]; columns: string[] } }>(`/backups/tables/${tableName}/preview`, {
      params: { limit }
    })
  },

  updateBackupNotes(id: number, notes: string) {
    return request.post(`/backups/${id}/notes`, { notes })
  }
}

export default backupApi
