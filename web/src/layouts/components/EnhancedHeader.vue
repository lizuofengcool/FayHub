<template>
  <header class="enhanced-header" :class="{ compact: props.compact }">
    <div class="header-left" v-if="!props.compact">
      <el-icon class="menu-toggle" @click="$emit('toggle-sidebar')">
        <Fold v-if="!sidebarCollapsed" />
        <Expand v-else />
      </el-icon>

      <el-breadcrumb separator="/">
        <el-breadcrumb-item :to="{ path: '/dashboard' }">
          <el-icon class="breadcrumb-home"><HomeFilled /></el-icon>
        </el-breadcrumb-item>
        <el-breadcrumb-item v-for="(item, index) in breadcrumbs" :key="index">
          {{ item }}
        </el-breadcrumb-item>
      </el-breadcrumb>
    </div>

    <div class="header-right">
      <el-tooltip content="搜索菜单 (Ctrl+K)" placement="bottom" v-if="!props.compact">
        <div class="header-action" @click="emit('open-search')">
          <el-icon><Search /></el-icon>
          <span class="action-text">搜索</span>
          <kbd class="shortcut">Ctrl K</kbd>
        </div>
      </el-tooltip>

      <el-tooltip :content="isDark ? '切换亮色' : '切换暗色'" placement="bottom" v-if="!props.compact">
        <div class="header-action" @click="toggleTheme">
          <el-icon><Sunny v-if="isDark" /><Moon v-else /></el-icon>
        </div>
      </el-tooltip>

      <el-tooltip :content="isFullscreen ? '退出全屏' : '全屏'" placement="bottom" v-if="!props.compact">
        <div class="header-action" @click="toggleFullscreen">
          <el-icon><FullScreen v-if="!isFullscreen" /><Aim v-else /></el-icon>
        </div>
      </el-tooltip>

      <el-tooltip content="通知" placement="bottom">
        <div class="header-action" @click="showNotifications = true">
          <el-badge :value="unreadCount" :hidden="unreadCount === 0" :max="99" class="notification-badge">
            <el-icon><Bell /></el-icon>
          </el-badge>
        </div>
      </el-tooltip>

      <el-tooltip content="布局设置" placement="bottom" v-if="!props.compact">
        <div class="header-action" @click="emit('open-settings')">
          <el-icon><Setting /></el-icon>
        </div>
      </el-tooltip>

      <el-dropdown trigger="click" @command="handleLangCommand" v-if="!props.compact">
        <div class="header-action">
          <el-icon><Document /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="zh-CN" :class="{ active: currentLang === 'zh-CN' }">
              简体中文
            </el-dropdown-item>
            <el-dropdown-item command="en-US" :class="{ active: currentLang === 'en-US' }">
              English
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>

      <!-- 用户信息 -->
      <el-dropdown trigger="hover" @command="handleUserCommand">
        <div class="user-info">
          <img :src="displayAvatar" alt="Avatar" class="user-avatar" @error="handleAvatarError">
          <span class="user-name">{{ userInfo.username }}</span>
          <el-icon class="user-arrow"><ArrowDown /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <div class="user-dropdown-header">
              <img :src="displayAvatar" alt="Avatar" class="dropdown-avatar" @error="handleAvatarError">
              <div class="dropdown-user-info">
                <div class="dropdown-username">{{ userInfo.username }}</div>
                <div class="dropdown-role">{{ displayRole }}</div>
              </div>
            </div>
            <el-dropdown-item divided command="profile">
              <el-icon><User /></el-icon>个人中心
            </el-dropdown-item>
            <el-dropdown-item command="settings">
              <el-icon><Setting /></el-icon>系统设置
            </el-dropdown-item>
            <el-dropdown-item divided command="logout">
              <el-icon><SwitchButton /></el-icon>退出登录
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <!-- 搜索弹窗 -->
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
          <el-icon><Search /></el-icon>
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
          <el-icon><Document /></el-icon>
          <div class="result-info">
            <div class="result-title">{{ result.title }}</div>
            <div class="result-path">{{ result.path }}</div>
          </div>
          <kbd class="result-shortcut" v-if="index < 9">{{ index + 1 }}</kbd>
        </div>
      </div>
      <div v-else-if="searchQuery" class="search-empty">
        <el-icon><InfoFilled /></el-icon>
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

    <!-- 通知弹窗 -->
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
        <el-icon :size="48"><Bell /></el-icon>
        <p>暂无通知</p>
      </div>
      <template #footer v-if="notifications.length > 0">
        <el-button text size="small" @click="clearAllNotifications">清空全部</el-button>
      </template>
    </el-dialog>

    <!-- 布局设置弹窗 -->
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
          <span class="setting-label">侧边栏折叠</span>
          <el-switch
            v-model="sidebarCollapsedLocal"
            active-text="折叠"
            inactive-text="展开"
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
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Fold, Expand, Search, FullScreen, Aim, Bell, ArrowDown,
  User, Setting, SwitchButton, Document, HomeFilled,
  Sunny, Moon, InfoFilled
} from '@element-plus/icons-vue'
import { useThemeStore } from '@/stores/theme'

interface Props {
  userInfo: any
  unreadCount: number
  currentPageTitle: string
  sidebarCollapsed?: boolean
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

const props = withDefaults(defineProps<Props>(), {
  sidebarCollapsed: false
})

const emit = defineEmits(['toggle-sidebar', 'logout', 'update:sidebarCollapsed', 'open-settings', 'open-search'])

const router = useRouter()
const route = useRoute()
const themeStore = useThemeStore()

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

// 布局设置
const isDark = computed({
  get: () => themeStore.isDark,
  set: (val) => themeStore.setTheme(val)
})
const sidebarCollapsedLocal = computed({
  get: () => props.sidebarCollapsed,
  set: (val) => emit('update:sidebarCollapsed', val)
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

// 面包屑
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

// 全屏切换
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

// 主题切换
const toggleTheme = () => {
  themeStore.toggleTheme()
}

// 搜索功能
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

// 键盘导航
const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'k' && (event.ctrlKey || event.metaKey)) {
    event.preventDefault()
    showSearch.value = true
    nextTick(() => {
      searchInputRef.value?.focus()
    })
  }

  if (!showSearch.value) return

  if (event.key === 'ArrowDown') {
    event.preventDefault()
    searchActiveIndex.value = (searchActiveIndex.value + 1) % searchResults.value.length
  } else if (event.key === 'ArrowUp') {
    event.preventDefault()
    searchActiveIndex.value = (searchActiveIndex.value - 1 + searchResults.value.length) % searchResults.value.length
  } else if (event.key === 'Enter') {
    event.preventDefault()
    const result = searchResults.value[searchActiveIndex.value]
    if (result) {
      navigateToSearchResult(result)
    }
  } else if (event.key === 'Escape' || event.keyCode === 27) {
    showSearch.value = false
  }
}

// 通知功能
const fetchNotifications = () => {
  notifications.value = [
    {
      id: 1,
      title: '系统更新',
      content: '系统已完成自动更新，版本 v2.5.0',
      time: new Date().toISOString(),
      read: false
    },
    {
      id: 2,
      title: '新用户注册',
      content: '有新用户注册，请及时审核',
      time: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
      read: false
    },
    {
      id: 3,
      title: '备份完成',
      content: '每日自动备份已完成',
      time: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(),
      read: true
    }
  ]
}

const markAsRead = (notification: Notification) => {
  notification.read = true
}

const clearAllNotifications = () => {
  notifications.value = []
}

const formatTime = (time: string) => {
  const now = new Date()
  const target = new Date(time)
  const diff = now.getTime() - target.getTime()

  if (diff < 1000 * 60) return '刚刚'
  if (diff < 1000 * 60 * 60) return `${Math.floor(diff / (1000 * 60))}分钟前`
  if (diff < 1000 * 60 * 60 * 24) return `${Math.floor(diff / (1000 * 60 * 60))}小时前`

  return target.toLocaleDateString()
}

const handleUserCommand = (command: string) => {
  switch (command) {
    case 'profile':
      router.push('/profile')
      break
    case 'settings':
      router.push('/system/settings')
      break
    case 'logout':
      emit('logout')
      break
  }
}

const handleLangCommand = (command: string) => {
  currentLang.value = command
  ElMessage.success(`已切换为 ${command === 'zh-CN' ? '简体中文' : 'English'}`)
}

const handleAvatarError = (e: Event) => {
  const img = e.target as HTMLImageElement
  img.src = 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 36 36"><circle cx="18" cy="18" r="18" fill="#e2e8f0"/><text x="18" y="23" text-anchor="middle" fill="#64748b" font-size="16" font-family="sans-serif">U</text></svg>')
}

onMounted(() => {
  fetchNotifications()
  document.addEventListener('keydown', handleKeydown)
  document.addEventListener('fullscreenchange', () => {
    isFullscreen.value = !!document.fullscreenElement
  })
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.enhanced-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 16px;
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
}

.enhanced-header.compact {
  height: auto;
  padding: 0;
  background: transparent;
  border-bottom: none;
  justify-content: flex-end;
}
.enhanced-header.compact .header-right {
  gap: 2px;
}
.enhanced-header.compact .user-info {
  margin-left: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.menu-toggle {
  font-size: 18px;
  color: #666;
  cursor: pointer;
  padding: 6px;
  border-radius: 6px;
  transition: all 0.2s;
}

.menu-toggle:hover {
  color: #18a058;
  background: rgba(24, 160, 88, 0.08);
}

.breadcrumb-home {
  font-size: 14px;
  margin-right: 2px;
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
  transition: all 0.2s;
  color: #666;
  font-size: 18px;
}

.header-action:hover {
  color: #18a058;
  background: rgba(24, 160, 88, 0.08);
}

.header-action .action-text {
  font-size: 13px;
}

.header-action .shortcut {
  font-size: 11px;
  padding: 1px 5px;
  background: #f0f0f0;
  border-radius: 3px;
  color: #999;
  border: 1px solid #e0e0e0;
}

.notification-badge :deep(.el-badge__content) {
  top: 6px;
  right: 6px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 10px;
  border-radius: 6px;
  transition: all 0.2s;
  margin-left: 4px;
}

.user-info:hover {
  background: rgba(24, 160, 88, 0.08);
}

.user-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 2px solid #e8e8e8;
}

.user-name {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

.user-arrow {
  font-size: 12px;
  color: #999;
}

/* 用户下拉菜单头部 */
.user-dropdown-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  gap: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.dropdown-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 2px solid #e8e8e8;
}

.dropdown-user-info {
  flex: 1;
}

.dropdown-username {
  font-size: 15px;
  font-weight: 600;
  color: #333;
}

.dropdown-role {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
}

/* 搜索弹窗 */
.search-dialog :deep(.el-dialog__body) {
  padding: 12px 20px 0;
}

.search-results {
  margin-top: 12px;
  max-height: 320px;
  overflow-y: auto;
}

.search-result-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
  margin-bottom: 2px;
}

.search-result-item:hover,
.search-result-item.active {
  background: rgba(24, 160, 88, 0.08);
}

.search-result-item.active .result-title {
  color: #18a058;
}

.result-info {
  flex: 1;
}

.result-title {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

.result-path {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
}

.result-shortcut {
  font-size: 11px;
  padding: 2px 6px;
  background: #f0f0f0;
  border-radius: 4px;
  color: #999;
  border: 1px solid #e0e0e0;
}

.search-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px 20px;
  color: #999;
  gap: 8px;
}

.search-empty .el-icon {
  font-size: 32px;
}

.search-footer {
  padding: 12px 0;
  border-top: 1px solid #f0f0f0;
  margin-top: 12px;
}

.search-hint {
  display: flex;
  gap: 16px;
  justify-content: center;
  font-size: 12px;
  color: #999;
}

.search-hint kbd {
  font-size: 11px;
  padding: 1px 5px;
  background: #f5f5f5;
  border-radius: 3px;
  border: 1px solid #e0e0e0;
}

/* 通知弹窗 */
.notification-dialog :deep(.el-dialog__body) {
  padding: 0;
}

.notification-list {
  max-height: 400px;
  overflow-y: auto;
}

.notification-item {
  display: flex;
  align-items: flex-start;
  padding: 14px 16px;
  border-bottom: 1px solid #f5f5f5;
  cursor: pointer;
  transition: all 0.2s;
  gap: 10px;
}

.notification-item:hover {
  background: #fafafa;
}

.notification-item.unread {
  background: #f0f9ff;
}

.notification-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #ccc;
  margin-top: 6px;
  flex-shrink: 0;
}

.notification-dot.unread {
  background: #18a058;
}

.notification-content-wrapper {
  flex: 1;
}

.notification-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.notification-content {
  font-size: 13px;
  color: #666;
  margin-top: 4px;
  line-height: 1.5;
}

.notification-time {
  font-size: 12px;
  color: #999;
  margin-top: 6px;
}

.notification-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 48px 20px;
  color: #ccc;
  gap: 12px;
}

.notification-empty p {
  font-size: 14px;
  color: #999;
}

/* 布局设置 */
.layout-settings {
  padding: 8px 0;
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 0;
  border-bottom: 1px solid #f5f5f5;
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-label {
  font-size: 14px;
  color: #333;
}
</style>
