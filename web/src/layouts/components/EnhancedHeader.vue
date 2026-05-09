<template>
  <header class="enhanced-header" :class="{ compact: props.compact }">
    <div class="header-left" v-if="!props.compact">
      <div class="custom-breadcrumb">
        <span class="breadcrumb-item breadcrumb-home" @click="router.push('/dashboard')">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>
        </span>
        <template v-for="(item, index) in breadcrumbs" :key="index">
          <span class="breadcrumb-separator">/</span>
          <span class="breadcrumb-item">{{ item }}</span>
        </template>
      </div>
    </div>

    <div class="header-right">
      <n-tooltip trigger="hover" v-if="!props.compact">
        <template #trigger>
          <div class="header-action" @click="emit('open-search')">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
            <span class="action-text">搜索</span>
            <kbd class="shortcut">Ctrl K</kbd>
          </div>
        </template>
        搜索菜单 (Ctrl+K)
      </n-tooltip>

      <n-tooltip trigger="hover" v-if="!props.compact">
        <template #trigger>
          <div class="header-action" @click="toggleTheme">
            <span v-if="isDark">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="5"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>
            </span>
            <span v-else>
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
            </span>
          </div>
        </template>
        {{ isDark ? '切换亮色' : '切换暗色' }}
      </n-tooltip>

      <n-tooltip trigger="hover" v-if="!props.compact">
        <template #trigger>
          <div class="header-action" @click="toggleFullscreen">
            <span v-if="!isFullscreen">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/></svg>
            </span>
            <span v-else>
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M8 3v3a2 2 0 0 1-2 2H3m18 0h-3a2 2 0 0 1-2-2V3m0 18v-3a2 2 0 0 1 2-2h3M3 16h3a2 2 0 0 1 2 2v3"/></svg>
            </span>
          </div>
        </template>
        {{ isFullscreen ? '退出全屏' : '全屏' }}
      </n-tooltip>

      <n-tooltip trigger="hover">
        <template #trigger>
          <div class="header-action" @click="showNotifications = true">
            <el-badge :value="unreadCount" :hidden="unreadCount === 0" :max="99" class="notification-badge">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>
            </el-badge>
          </div>
        </template>
        通知
      </n-tooltip>

      <n-tooltip trigger="hover" v-if="!props.compact">
        <template #trigger>
          <div class="header-action" @click="emit('open-settings')">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
          </div>
        </template>
        布局设置
      </n-tooltip>

      <n-dropdown trigger="click" :options="langOptions" @select="handleLangCommand" v-if="!props.compact">
        <div class="header-action">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
        </div>
      </n-dropdown>

      <n-dropdown trigger="hover" :options="userDropdownOptions" @select="handleUserCommand">
        <div class="user-info">
          <img :src="displayAvatar" alt="Avatar" class="user-avatar" @error="handleAvatarError">
          <span class="user-name">{{ userInfo.username }}</span>
          <svg class="user-arrow" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="m6 9 6 6 6-6"/></svg>
        </div>
      </n-dropdown>
    </div>

    <!-- 搜索弹窗 (Element Plus 复杂组件保留) -->
    <el-dialog v-model="showSearch" title="搜索" width="560px" :show-close="false" class="search-dialog">
      <el-input
        v-model="searchQuery"
        placeholder="搜索菜单、页面..."
        size="large"
        clearable
        @input="handleSearch"
        ref="searchInputRef"
      >
        <template #prefix>
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
        </template>
      </el-input>
      <div v-if="searchResults.length > 0" class="search-results">
        <div
          v-for="(result, index) in searchResults"
          :key="result.id"
          class="search-result-item"
          :class="{ active: searchActiveIndex === index }"
          @click="navigateToSearchResult(result)"
          @mouseenter="searchActiveIndex = index"
        >
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
          <div class="result-info">
            <div class="result-title">{{ result.title }}</div>
            <div class="result-path">{{ result.path }}</div>
          </div>
          <kbd class="result-shortcut" v-if="index < 9">{{ index + 1 }}</kbd>
        </div>
      </div>
      <div v-else-if="searchQuery" class="search-empty">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z"/></svg>
        <span>未找到相关结果</span>
      </div>
      <div class="search-footer">
        <div class="search-hint">
          <span><kbd>↑</kbd> <kbd>↓</kbd> 选择</span>
          <span><kbd>Enter</kbd> 确认</span>
          <span><kbd>Esc</kbd> 关闭</span>
        </div>
      </div>
    </el-dialog>

    <!-- 通知弹窗 (Element Plus 复杂组件保留) -->
    <el-dialog v-model="showNotifications" title="通知中心" width="400px" class="notification-dialog">
      <div v-if="notifications.length > 0" class="notification-list">
        <div
          v-for="notification in notifications"
          :key="notification.id"
          class="notification-item"
          :class="{ unread: !notification.read }"
          @click="markAsRead(notification)"
        >
          <div class="notification-dot" :class="{ unread: !notification.read }"></div>
          <div class="notification-content-wrapper">
            <div class="notification-title">{{ notification.title }}</div>
            <div class="notification-content">{{ notification.content }}</div>
            <div class="notification-time">{{ formatTime(notification.time) }}</div>
          </div>
        </div>
      </div>
      <div v-else class="notification-empty">
        <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>
        <p>暂无通知</p>
      </div>
      <template #footer v-if="notifications.length > 0">
        <el-button text size="small" @click="clearAllNotifications">清空全部</el-button>
      </template>
    </el-dialog>

    <!-- 布局设置弹窗 (Element Plus 复杂组件保留) -->
    <el-drawer v-model="showLayoutSettings" title="布局设置" size="300px">
      <div class="layout-settings">
        <div class="setting-item">
          <span class="setting-label">主题模式</span>
          <el-switch
            v-model="isDark"
            active-text="暗色"
            inactive-text="亮色"
            @change="toggleTheme"
          />
        </div>
        <div class="setting-item">
          <span class="setting-label">显示标签栏</span>
          <el-switch v-model="showTabs" active-text="显示" inactive-text="隐藏" />
        </div>
        <div class="setting-item">
          <span class="setting-label">固定头部</span>
          <el-switch v-model="fixedHeader" active-text="固定" inactive-text="不固定" />
        </div>
      </div>
    </el-drawer>
  </header>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, h } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMessage, NIcon, NTooltip, NDropdown } from 'naive-ui'
import { useThemeStore } from '@/stores/theme'

interface Props {
  userInfo: any
  unreadCount: number
  currentPageTitle: string
  compact?: boolean
}

interface Notification {
  id: number
  title: string
  content: string
  time: string
  read: boolean
}

interface SearchResult {
  id: string
  title: string
  path: string
}

const props = withDefaults(defineProps<Props>(), {})

const emit = defineEmits(['logout', 'open-settings', 'open-search'])

const router = useRouter()
const route = useRoute()
const themeStore = useThemeStore()
const message = useMessage()

const isFullscreen = ref(false)
const searchQuery = ref('')
const searchResults = ref<SearchResult[]>([])
const searchActiveIndex = ref(0)
const notifications = ref<Notification[]>([])
const showSearch = ref(false)
const showNotifications = ref(false)
const showLayoutSettings = ref(false)
const searchInputRef = ref()
const currentLang = ref('zh-CN')

const isDark = computed({
  get: () => themeStore.isDark,
  set: (val) => themeStore.setTheme(val)
})
const showTabs = ref(true)
const fixedHeader = ref(true)

const displayAvatar = computed(() => {
  return props.userInfo.avatar || 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 36 36"><circle cx="18" cy="18" r="18" fill="#e2e8f0"/><text x="18" y="23" text-anchor="middle" fill="#64748b" font-size="16" font-family="sans-serif">' + (props.userInfo.username?.[0] || 'U') + '</text></svg>')
})

const displayRole = computed(() => {
  const roleMap: Record<string, string> = {
    'super_admin': '超级管理员',
    'tenant_admin': '租户管理员',
    'user': '普通用户'
  }
  return roleMap[props.userInfo.role] || props.userInfo.role || '普通用户'
})

const breadcrumbs = computed(() => {
  const matched = route.matched
  const items: string[] = []
  for (const record of matched) {
    if (record.meta?.title && record.path !== '/dashboard') {
      items.push(record.meta.title as string)
    }
  }
  if (items.length === 0 && props.currentPageTitle && props.currentPageTitle !== '首页') {
    items.push(props.currentPageTitle)
  }
  return items
})

const renderIcon = (iconName: string) => {
  return () => h(NIcon, null, {
    default: () => h('i', { class: `ri-${iconName}` })
  })
}

const langOptions = computed(() => [
  { label: '简体中文', key: 'zh-CN' },
  { label: 'English', key: 'en-US' }
])

const userDropdownOptions = computed(() => [
  {
    key: 'header',
    type: 'render' as const,
    render: () => h('div', { class: 'user-dropdown-header' }, [
      h('img', { src: displayAvatar.value, alt: 'Avatar', class: 'dropdown-avatar', onError: handleAvatarError }),
      h('div', { class: 'dropdown-user-info' }, [
        h('div', { class: 'dropdown-username' }, props.userInfo.username),
        h('div', { class: 'dropdown-role' }, displayRole.value)
      ])
    ])
  },
  { type: 'divider' as const, key: 'd1' },
  { label: '个人中心', key: 'profile', icon: renderIcon('user-line') },
  { label: '系统设置', key: 'settings', icon: renderIcon('settings-line') },
  { type: 'divider' as const, key: 'd2' },
  { label: '退出登录', key: 'logout', icon: renderIcon('logout-box-r-line') }
])

const toggleFullscreen = () => {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen().then(() => {
      isFullscreen.value = true
    }).catch(() => {})
  } else {
    document.exitFullscreen().then(() => {
      isFullscreen.value = false
    }).catch(() => {})
  }
}

const toggleTheme = () => {
  themeStore.toggleTheme()
}

const handleSearch = () => {
  if (!searchQuery.value.trim()) {
    searchResults.value = []
    return
  }

  const query = searchQuery.value.toLowerCase()
  const mockResults: SearchResult[] = [
    { id: '1', title: '仪表盘', path: '/dashboard' },
    { id: '2', title: '用户管理', path: '/system/user' },
    { id: '3', title: '角色权限', path: '/system/role' },
    { id: '4', title: '菜单管理', path: '/system/menu' },
    { id: '5', title: '系统设置', path: '/system/settings' },
    { id: '6', title: '租户管理', path: '/system/tenant' },
    { id: '7', title: 'API 管理', path: '/system/api' },
    { id: '8', title: '文件管理', path: '/system/file' }
  ]

  searchResults.value = mockResults.filter(item =>
    item.title.toLowerCase().includes(query) ||
    item.path.toLowerCase().includes(query)
  )
  searchActiveIndex.value = 0
}

const navigateToSearchResult = (result: SearchResult) => {
  router.push(result.path)
  searchQuery.value = ''
  searchResults.value = []
  showSearch.value = false
}

const handleAvatarError = (e: Event) => {
  const img = e.target as HTMLImageElement
  img.src = 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 36 36"><circle cx="18" cy="18" r="18" fill="#e2e8f0"/><text x="18" y="23" text-anchor="middle" fill="#64748b" font-size="16" font-family="sans-serif">' + (props.userInfo.username?.[0] || 'U') + '</text></svg>')
}

const handleUserCommand = (key: string) => {
  switch (key) {
    case 'profile':
      router.push('/profile')
      break
    case 'settings':
      emit('open-settings')
      break
    case 'logout':
      emit('logout')
      break
  }
}

const handleLangCommand = (key: string) => {
  currentLang.value = key
  message.success(key === 'zh-CN' ? '已切换为简体中文' : 'Switched to English')
}

const markAsRead = (notification: Notification) => {
  notification.read = true
}

const clearAllNotifications = () => {
  notifications.value = []
  message.success('已清空所有通知')
}

const formatTime = (time: string) => {
  return time
}

onMounted(() => {
  notifications.value = [
    { id: 1, title: '系统更新通知', content: 'FayHub v2.5.0 已发布，包含多项性能优化', time: '10分钟前', read: false },
    { id: 2, title: '安全提醒', content: '检测到异常登录尝试，请检查账户安全', time: '1小时前', read: false },
    { id: 3, title: '任务完成', content: '数据备份任务已成功完成', time: '3小时前', read: true },
  ]
})

onBeforeUnmount(() => {})
</script>

<style scoped>
.enhanced-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: var(--nav-height, 56px);
  padding: 0 20px;
  background: var(--card-bg, #fff);
  border-bottom: 1px solid var(--border-color, #e8e8e8);
  flex-shrink: 0;
}
.enhanced-header.compact {
  height: 44px;
  padding: 0 12px;
}

.header-left {
  display: flex;
  align-items: center;
}

.custom-breadcrumb {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}

.breadcrumb-item {
  color: var(--text-secondary, #666);
}

.breadcrumb-home {
  cursor: pointer;
  color: var(--text-muted, #999);
  display: flex;
  align-items: center;
}
.breadcrumb-home:hover {
  color: var(--primary, #4f46e5);
}

.breadcrumb-separator {
  color: var(--text-muted, #ccc);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 4px;
}

.header-action {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  border-radius: 6px;
  cursor: pointer;
  color: var(--text-secondary, #666);
  transition: all 0.15s;
  font-size: 13px;
}
.header-action:hover {
  background: var(--hover-bg, rgba(0,0,0,0.06));
  color: var(--text-primary, #333);
}

.action-text {
  font-size: 13px;
}

.shortcut {
  padding: 2px 6px;
  font-size: 11px;
  background: var(--hover-bg, rgba(0,0,0,0.06));
  border-radius: 4px;
  color: var(--text-muted, #999);
  font-family: inherit;
}

.notification-badge {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s;
}
.user-info:hover {
  background: var(--hover-bg, rgba(0,0,0,0.06));
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary, #333);
}

.user-arrow {
  color: var(--text-muted, #999);
}

.user-dropdown-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
}

.dropdown-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
}

.dropdown-user-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.dropdown-username {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #333);
}

.dropdown-role {
  font-size: 12px;
  color: var(--text-muted, #999);
}

.search-dialog .search-results {
  margin-top: 12px;
  max-height: 300px;
  overflow-y: auto;
}

.search-result-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
}
.search-result-item:hover,
.search-result-item.active {
  background: var(--hover-bg, rgba(0,0,0,0.06));
}

.result-info {
  flex: 1;
}

.result-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary, #333);
}

.result-path {
  font-size: 12px;
  color: var(--text-muted, #999);
}

.result-shortcut {
  padding: 2px 6px;
  font-size: 11px;
  background: var(--hover-bg, rgba(0,0,0,0.06));
  border-radius: 4px;
  color: var(--text-muted, #999);
}

.search-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px;
  color: var(--text-muted, #999);
}

.search-footer {
  margin-top: 12px;
  padding-top: 8px;
  border-top: 1px solid var(--border-color, #e8e8e8);
}

.search-hint {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: var(--text-muted, #999);
}
.search-hint kbd {
  padding: 1px 4px;
  font-size: 11px;
  background: var(--hover-bg, rgba(0,0,0,0.06));
  border-radius: 3px;
  font-family: inherit;
}

.notification-dialog .notification-list {
  max-height: 400px;
  overflow-y: auto;
}

.notification-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
}
.notification-item:hover {
  background: var(--hover-bg, rgba(0,0,0,0.04));
}
.notification-item.unread {
  background: var(--primary-suppl, rgba(79,70,229,0.04));
}

.notification-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: transparent;
  margin-top: 6px;
  flex-shrink: 0;
}
.notification-dot.unread {
  background: var(--primary, #4f46e5);
}

.notification-content-wrapper {
  flex: 1;
}

.notification-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary, #333);
}

.notification-content {
  font-size: 13px;
  color: var(--text-secondary, #666);
  margin-top: 4px;
}

.notification-time {
  font-size: 12px;
  color: var(--text-muted, #999);
  margin-top: 4px;
}

.notification-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 32px;
  color: var(--text-muted, #999);
}

.layout-settings {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 16px 0;
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.setting-label {
  font-size: 14px;
  color: var(--text-primary, #333);
}
</style>