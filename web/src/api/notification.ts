import request, { type ApiResponse } from './request'

export interface Notification {
  id: number
  title: string
  content: string
  type: string
  category: string
  is_read: boolean
  sender_name?: string
  created_at: string
}

export interface NotificationChannel {
  id: number
  tenant_id: number
  name: string
  type: string
  provider: string
  config: string
  status: number
  is_default: boolean
  created_at: string
  updated_at: string
}

export interface NotificationTemplate {
  id: number
  tenant_id: number
  name: string
  code: string
  channel_id: number
  subject: string
  content: string
  status: number
  created_at: string
  updated_at: string
}

export interface NotificationRecord {
  id: number
  tenant_id: number
  channel_id: number
  template_id: number
  recipient: string
  subject: string
  content: string
  status: string
  error: string
  sent_at: string | null
  created_at: string
}

export interface ListResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

const notificationApi = {
  listNotifications(params?: Record<string, any>): Promise<ApiResponse<ListResult<Notification>>> {
    return request.get('/notifications', { params })
  },

  getUnreadCount(): Promise<ApiResponse<{ unread_count: number }>> {
    return request.get('/notifications/unread-count')
  },

  markAsRead(ids: number[]): Promise<ApiResponse<null>> {
    return request.put('/notifications/mark-read', { ids })
  },

  markAllAsRead(): Promise<ApiResponse<null>> {
    return request.put('/notifications/mark-all-read')
  },

  deleteNotifications(ids: number[]): Promise<ApiResponse<null>> {
    return request.delete('/notifications', { data: { ids } })
  },

  listChannels(page = 1, pageSize = 20): Promise<ApiResponse<ListResult<NotificationChannel>>> {
    return request.get('/notification-channels', { params: { page, page_size: pageSize } })
  },

  getChannel(id: number): Promise<ApiResponse<NotificationChannel>> {
    return request.get(`/notification-channels/${id}`)
  },

  createChannel(data: Partial<NotificationChannel>): Promise<ApiResponse<NotificationChannel>> {
    return request.post('/notification-channels', data)
  },

  updateChannel(id: number, data: Partial<NotificationChannel>): Promise<ApiResponse<NotificationChannel>> {
    return request.put(`/notification-channels/${id}`, data)
  },

  deleteChannel(id: number): Promise<ApiResponse<null>> {
    return request.delete(`/notification-channels/${id}`)
  },

  listTemplates(page = 1, pageSize = 20): Promise<ApiResponse<ListResult<NotificationTemplate>>> {
    return request.get('/notification-templates', { params: { page, page_size: pageSize } })
  },

  createTemplate(data: Partial<NotificationTemplate>): Promise<ApiResponse<NotificationTemplate>> {
    return request.post('/notification-templates', data)
  },

  updateTemplate(id: number, data: Partial<NotificationTemplate>): Promise<ApiResponse<NotificationTemplate>> {
    return request.put(`/notification-templates/${id}`, data)
  },

  deleteTemplate(id: number): Promise<ApiResponse<null>> {
    return request.delete(`/notification-templates/${id}`)
  },

  getRecords(page = 1, pageSize = 20): Promise<ApiResponse<ListResult<NotificationRecord>>> {
    return request.get('/notification-records', { params: { page, page_size: pageSize } })
  },

  send(data: Partial<NotificationRecord>): Promise<ApiResponse<null>> {
    return request.post('/notification-records/send', data)
  }
}

export default notificationApi
