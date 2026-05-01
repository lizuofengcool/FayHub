import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import authApi, { type UserInfo, type LoginParams } from '@/api/auth'

function getTokenFromCookie(): string {
  const match = document.cookie.match(/(?:^|;\s*)fayhub_token=([^;]*)/)
  return match ? decodeURIComponent(match[1]) : ''
}

function getTokenFromStorage(): string {
  return localStorage.getItem('fayhub_token') || getTokenFromCookie()
}

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(getTokenFromStorage())
  const userInfo = ref<UserInfo | null>(null)

  const isLoggedIn = computed(() => !!token.value)
  const isSuperAdmin = computed(() => userInfo.value?.role === 'super_admin')
  const isTenantAdmin = computed(() => userInfo.value?.role === 'tenant_admin')
  const tenantId = computed(() => userInfo.value?.tenant_id || 0)
  const displayName = computed(() => userInfo.value?.username || '')
  const role = computed(() => {
    const r = userInfo.value?.role
    if (r === 'super_admin') return '超级管理员'
    if (r === 'tenant_admin') return '租户管理员'
    return '普通用户'
  })

  async function login(params: LoginParams) {
    const res = await authApi.login(params)
    const data = res.data
    token.value = data.token
    localStorage.setItem('fayhub_token', data.token)

    userInfo.value = {
      id: data.user_id,
      user_id: data.user_id,
      username: data.username,
      role: data.role,
      tenant_id: data.tenant_id
    }
    localStorage.setItem('userInfo', JSON.stringify(userInfo.value))

    return data
  }

  async function fetchCurrentUser() {
    const res = await authApi.getCurrentUser()
    userInfo.value = res.data
    localStorage.setItem('userInfo', JSON.stringify(res.data))
    return res.data
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch {}
    token.value = ''
    userInfo.value = null
    document.cookie = 'fayhub_token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;'
    localStorage.removeItem('fayhub_token')
    localStorage.removeItem('userInfo')
  }

  function loadFromStorage() {
    token.value = getTokenFromStorage()
    const stored = localStorage.getItem('userInfo')
    if (stored) {
      try {
        userInfo.value = JSON.parse(stored)
      } catch {
        userInfo.value = null
      }
    }
  }

  loadFromStorage()

  return {
    token,
    userInfo,
    isLoggedIn,
    isSuperAdmin,
    isTenantAdmin,
    tenantId,
    displayName,
    role,
    login,
    fetchCurrentUser,
    logout,
    loadFromStorage
  }
})
