import request, { type ApiResponse } from './request'

export interface OnlineUser {
  user_id: number
  username: string
  nickname: string
  email: string
  role: string
  tenant_id: number
  ip: string
  user_agent: string
  login_at: string
  last_seen: string
}

const onlineUserApi = {
  getOnlineUsers(): Promise<ApiResponse<OnlineUser[]>> {
    return request.get('/online-users')
  },

  forceLogout(userId: number): Promise<ApiResponse<null>> {
    return request.post('/online-users/force-logout', { user_id: userId })
  },

  getOnlineCount(): Promise<ApiResponse<{ count: number }>> {
    return request.get('/online-users/count')
  }
}

export default onlineUserApi
