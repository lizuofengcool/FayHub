import request, { type PageParams, type PageResult } from './request'

export interface Notification {
  id: number
  user_id: number
  title: string
  content: string
  type: string
  category: string
  is_read: boolean
  read_at: string | null
  data: Record<string, any>
  sender_id: number
  sender_name: string
  link: string
  priority: number
  expired_at: string | null
  created_at: string
}

export interface NotificationFilters extends PageParams {
  type?: string
  category?: string
  is_read?: boolean
}

export interface SendNotificationRequest {
  user_ids: number[]
  title: string
  content: string
  type?: string
  category?: string
  data?: Record<string, any>
  sender_id?: number
  sender_name?: string
  link?: string
  priority?: number
}

const notificationApi = {
  listNotifications(params?: NotificationFilters) {
    return request.get<any, { code: number; data: PageResult<Notification> }>('/notifications', { params })
  },

  getUnreadCount() {
    return request.get<any, { code: number; data: { unread_count: number } }>('/notifications/unread-count')
  },

  markAsRead(ids: number[]) {
    return request.put('/notifications/read', { ids })
  },

  markAllAsRead() {
    return request.put('/notifications/read-all')
  },

  deleteNotifications(ids: number[]) {
    return request.delete('/notifications', { data: { ids } })
  },

  sendNotification(data: SendNotificationRequest) {
    return request.post('/notifications/send', data)
  }
}

export default notificationApi
