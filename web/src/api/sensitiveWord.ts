import request, { type PageResult } from './request'

export interface SensitiveWord {
  id: number
  word: string
  category: string
  level: number
  status: number
  created_at: string
  updated_at: string
}

export interface SensitiveWordParams {
  page?: number
  page_size?: number
  keyword?: string
  category?: string
  level?: number
}

const sensitiveWordApi = {
  list(params?: SensitiveWordParams) {
    return request.get<any, { code: number; data: PageResult<SensitiveWord> }>('/sensitive-words', { params })
  },

  create(data: { word: string; category?: string; level?: number }) {
    return request.post('/sensitive-words', data)
  },

  update(id: number, data: Partial<SensitiveWord>) {
    return request.put(`/sensitive-words/${id}`, data)
  },

  delete(id: number) {
    return request.delete(`/sensitive-words/${id}`)
  },

  batchCreate(data: { words: string[]; category?: string; level?: number }) {
    return request.post('/sensitive-words/batch', data)
  },

  rebuild() {
    return request.post('/sensitive-words/rebuild')
  },

  check(text: string) {
    return request.post('/sensitive-words/check', { text })
  }
}

export default sensitiveWordApi
