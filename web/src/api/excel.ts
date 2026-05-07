import request from '@/utils/request'

export function importFile(formData: FormData) {
  return request({
    url: '/api/excel/import',
    method: 'post',
    data: formData,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function previewFile(formData: FormData) {
  return request({
    url: '/api/excel/preview',
    method: 'post',
    data: formData,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function downloadTemplate() {
  return request({
    url: '/api/excel/template',
    method: 'get',
    responseType: 'blob'
  })
}

export function exportGeneric(params: { format?: string; prefix?: string }) {
  return request({
    url: '/api/excel/export',
    method: 'get',
    params,
    responseType: 'blob'
  })
}

export function exportData(data: {
  format?: string
  prefix?: string
  columns: { header: string; field: string }[]
  rows: Record<string, any>[]
  sheetName?: string
}) {
  return request({
    url: '/api/excel/export',
    method: 'post',
    data,
    responseType: 'blob'
  })
}
