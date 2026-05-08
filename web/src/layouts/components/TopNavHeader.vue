<template>
  <header class="top-nav-header">
    <div class="top-nav-left">
      <div class="top-nav-logo" v-if="prefsStore.prefs.logo">
        <div class="logo-icon">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="18" height="18" rx="3"/>
            <path d="M3 9h18M9 21V9"/>
          </svg>
        </div>
        <span class="logo-text">FayHub</span>
      </div>

      <nav class="top-nav-menu">
        <div
          v-for="menu in menuItems"
          :key="menu.id"
          class="top-nav-item"
          :class="{ active: isMenuActive(menu) }"
          @mouseenter="mode === 'hover' && handleMenuHover(menu)"
          @mouseleave="mode === 'hover' && handleMenuLeave(menu)"
          @click="handleMenuClick(menu)"
        >
          <span class="top-nav-label">{{ menu.title }}</span>
          <svg v-if="menu.children && menu.children.length" class="top-nav-arrow" viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2">
            <path d="m6 9 6 6 6-6"/>
          </svg>

          <div
            v-if="mode === 'hover' && menu.children && menu.children.length && hoveredMenuId === menu.id"
            class="top-nav-dropdown"
            @mouseenter="handleDropdownEnter(menu)"
            @mouseleave="handleDropdownLeave(menu)"
          >
            <div class="dropdown-inner">
              <div class="dropdown-left" :class="{ 'dropdown-left-full': !hasAnyGrandchild(menu) }">
                <div
                  v-for="child in menu.children"
                  :key="child.id"
                  class="dropdown-item"
                  :class="{ active: activeChildId === child.id, 'has-children': child.children && child.children.length }"
                  @mouseenter="activeChildId = child.id"
                  @click.stop="handleChildClick(child)"
                >
                  <span>{{ child.title }}</span>
                  <svg v-if="child.children && child.children.length" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="m9 18 6-6-6-6"/>
                  </svg>
                </div>
              </div>
              <div class="dropdown-right" v-if="activeChildChildren.length">
                <div class="dropdown-right-title">{{ activeChildTitle }}</div>
                <div
                  v-for="grandchild in activeChildChildren"
                  :key="grandchild.id"
                  class="dropdown-sub-item"
                  @click.stop="handleGrandchildClick(grandchild)"
                >
                  {{ grandchild.title }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </nav>
    </div>

    <div class="top-nav-right">
      <div class="header-action" title="搜索 (Ctrl+K)" @click="emit('open-search')">
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
        </svg>
      </div>

      <el-tooltip :content="themeStore.isDark ? '切换亮色' : '切换暗色'" placement="bottom">
        <div class="header-action" @click="toggleTheme">
          <svg v-if="themeStore.isDark" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="5"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
          </svg>
          <svg v-else viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
          </svg>
        </div>
      </el-tooltip>

      <div
        class="header-action notification-wrapper"
        title="通知"
        @mouseenter="showNotifications = true"
        @mouseleave="handleNotificationLeave"
      >
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/>
        </svg>
        <span v-if="unreadCount" class="badge-dot">{{ unreadCount > 99 ? '99+' : unreadCount }}</span>

        <div
          v-if="showNotifications"
          class="notification-popup"
          @mouseenter="showNotifications = true"
          @mouseleave="showNotifications = false"
        >
          <div class="notification-popup-header">
            <span>通知中心</span>
            <span class="notification-clear" @click.stop="handleClearNotifications">清空</span>
          </div>
          <div class="notification-popup-body">
            <div v-if="notifications.length > 0" class="notification-list">
              <div
                v-for="item in notifications"
                :key="item.id"
                class="notification-item"
                :class="{ unread: !item.read }"
                @click.stop="handleNotificationClick(item)"
              >
                <div class="notification-dot" :class="{ unread: !item.read }"></div>
                <div class="notification-content-wrapper">
                  <div class="notification-title">{{ item.title }}</div>
                  <div class="notification-content">{{ item.content }}</div>
                  <div class="notification-time">{{ item.time }}</div>
                </div>
              </div>
            </div>
            <div v-else class="notification-empty">
              <p>暂无通知</p>
            </div>
          </div>
        </div>
      </div>

      <div class="header-action" title="布局设置" @click="emit('open-settings')">
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
        </svg>
      </div>

      <div class="user-info" @click="showUserMenu = !showUserMenu" v-click-outside="() => showUserMenu = false">
        <img :src="userInfo.avatar || defaultAvatar" alt="Avatar" class="user-avatar">
        <span class="user-name">{{ userInfo.username }}</span>

        <div v-if="showUserMenu" class="user-dropdown">
          <div class="user-dropdown-item" @click.stop="handleProfile">个人中心</div>
          <div class="user-dropdown-item" @click.stop="handleSettings">系统设置</div>
          <div class="user-dropdown-divider"></div>
          <div class="user-dropdown-item" @click.stop="emit('logout')">退出登录</div>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { usePreferencesStore } from '@/stores/preferences'
import { useThemeStore } from '@/stores/theme'

interface MenuItem {
  id: number
  title: string
  path?: string
  children?: MenuItem[]
}

interface NotificationItem {
  id: number
  title: string
  content: string
  time: string
  read: boolean
}

const props = withDefaults(defineProps<{
  menuItems: MenuItem[]
  userInfo: { id: number; username: string; role?: string; avatar?: string }
  unreadCount: number
  currentPageTitle: string
  mode?: 'hover' | 'click'
}>(), {
  mode: 'hover'
})

const emit = defineEmits<{
  'open-search': []
  'open-settings': []
  logout: []
  'menu-select': [menu: MenuItem]
}>()

const router = useRouter()
const route = useRoute()
const prefsStore = usePreferencesStore()
const themeStore = useThemeStore()

const hoveredMenuId = ref<number | null>(null)
const activeChildId = ref<number | null>(null)
const showUserMenu = ref(false)
const showNotifications = ref(false)
const hoverTimer = ref<ReturnType<typeof setTimeout> | null>(null)
const leaveTimer = ref<ReturnType<typeof setTimeout> | null>(null)
const notificationLeaveTimer = ref<ReturnType<typeof setTimeout> | null>(null)

const notifications = ref<NotificationItem[]>([
  { id: 1, title: '系统更新通知', content: 'FayHub v2.5.0 已发布，包含多项性能优化', time: '10分钟前', read: false },
  { id: 2, title: '安全提醒', content: '检测到异常登录尝试，请检查账户安全', time: '1小时前', read: false },
  { id: 3, title: '任务完成', content: '数据备份任务已成功完成', time: '3小时前', read: true },
])

const defaultAvatar = 'https://api.dicebear.com/7.x/avataaars/svg?seed=admin&backgroundColor=e2e8f0'

function toggleTheme() {
  themeStore.toggleTheme()
}

function handleNotificationLeave() {
  notificationLeaveTimer.value = setTimeout(() => {
    showNotifications.value = false
  }, 200)
}

function handleClearNotifications() {
  notifications.value = []
}

function handleNotificationClick(item: NotificationItem) {
  item.read = true
}

function isMenuActive(menu: MenuItem): boolean {
  if (menu.children) {
    return menu.children.some((child: MenuItem) =>
      route.path === child.path || route.path.startsWith(child.path + '/')
    )
  }
  return route.path === menu.path
}

function handleMenuHover(menu: MenuItem) {
  if (leaveTimer.value) { clearTimeout(leaveTimer.value); leaveTimer.value = null }
  if (!menu.children || !menu.children.length) return
  hoverTimer.value = setTimeout(() => {
    hoveredMenuId.value = menu.id
    if (menu.children && menu.children.length) {
      activeChildId.value = menu.children[0].id
    }
  }, 150)
}

function handleMenuLeave(_menu: MenuItem) {
  if (hoverTimer.value) { clearTimeout(hoverTimer.value); hoverTimer.value = null }
  leaveTimer.value = setTimeout(() => {
    hoveredMenuId.value = null
    activeChildId.value = null
  }, 200)
}

function handleDropdownEnter(_menu: MenuItem) {
  if (leaveTimer.value) { clearTimeout(leaveTimer.value); leaveTimer.value = null }
}

function handleDropdownLeave(_menu: MenuItem) {
  leaveTimer.value = setTimeout(() => {
    hoveredMenuId.value = null
    activeChildId.value = null
  }, 200)
}

function handleMenuClick(menu: MenuItem) {
  if (props.mode === 'click') {
    emit('menu-select', menu)
    return
  }
  if (!menu.children || !menu.children.length) {
    if (menu.path) router.push(menu.path)
  }
}

function hasAnyGrandchild(menu: MenuItem): boolean {
  if (!menu.children) return false
  return menu.children.some(child => child.children && child.children.length > 0)
}

function handleChildClick(child: MenuItem) {
  if (!child.children || !child.children.length) {
    if (child.path) router.push(child.path)
    hoveredMenuId.value = null
    activeChildId.value = null
  }
}

function handleGrandchildClick(grandchild: MenuItem) {
  if (grandchild.path) router.push(grandchild.path)
  hoveredMenuId.value = null
  activeChildId.value = null
}

const activeChildTitle = computed(() => {
  if (!hoveredMenuId.value || !activeChildId.value) return ''
  const menu = props.menuItems.find(m => m.id === hoveredMenuId.value)
  if (!menu?.children) return ''
  const child = menu.children.find(c => c.id === activeChildId.value)
  return child?.title || ''
})

const activeChildChildren = computed(() => {
  if (!hoveredMenuId.value || !activeChildId.value) return []
  const menu = props.menuItems.find(m => m.id === hoveredMenuId.value)
  if (!menu?.children) return []
  const child = menu.children.find(c => c.id === activeChildId.value)
  return child?.children || []
})

function handleProfile() {
  showUserMenu.value = false
  router.push('/profile')
}

function handleSettings() {
  showUserMenu.value = false
  router.push('/settings')
}
</script>

<style scoped>
.top-nav-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: var(--nav-height, 56px);
  background: var(--header-bg, #fff);
  border-bottom: 1px solid var(--border-color, #e8e8e8);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  z-index: 50;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.top-nav-left {
  display: flex;
  align-items: center;
  gap: 24px;
  flex: 1;
  min-width: 0;
}

.top-nav-logo {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}
.logo-icon {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background: var(--primary, #2d8cf0);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}
.logo-text {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary, #333);
}

.top-nav-menu {
  display: flex;
  align-items: center;
  gap: 4px;
  height: 100%;
}

.top-nav-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 0 14px;
  height: 100%;
  cursor: pointer;
  font-size: 14px;
  color: var(--text-secondary, #666);
  transition: all 0.15s;
  white-space: nowrap;
}
.top-nav-item:hover {
  color: var(--primary, #2d8cf0);
  background: rgba(45, 140, 240, 0.04);
}
.top-nav-item.active {
  color: var(--primary, #2d8cf0);
  background: var(--primary-suppl, rgba(45, 140, 240, 0.08));
}
.top-nav-item.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 14px;
  right: 14px;
  height: 2px;
  background: var(--primary, #2d8cf0);
  border-radius: 1px 1px 0 0;
}

.top-nav-arrow {
  opacity: 0.5;
  transition: transform 0.15s;
}
.top-nav-item:hover .top-nav-arrow {
  opacity: 1;
}

.top-nav-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  min-width: 480px;
  background: var(--card-bg, #fff);
  border: 1px solid var(--border-color, #e8e8e8);
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
  z-index: 100;
  animation: dropdownIn 0.15s ease-out;
}

@keyframes dropdownIn {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}

.dropdown-inner {
  display: flex;
  min-height: 160px;
}

.dropdown-left {
  width: 180px;
  padding: 8px 0;
  border-right: 1px solid var(--border-color, #e8e8e8);
  background: var(--body-bg, #fafafa);
  border-radius: 8px 0 0 8px;
}
.dropdown-left-full {
  width: 100%;
  border-right: none;
  border-radius: 8px;
}

.dropdown-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  font-size: 14px;
  color: var(--text-secondary, #666);
  cursor: pointer;
  transition: all 0.12s;
}
.dropdown-item:hover,
.dropdown-item.active {
  color: var(--primary, #2d8cf0);
  background: rgba(45, 140, 240, 0.06);
}

.dropdown-right {
  flex: 1;
  padding: 12px 16px;
}

.dropdown-right-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary, #333);
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color, #eee);
}

.dropdown-sub-item {
  padding: 8px 12px;
  font-size: 13px;
  color: var(--text-secondary, #666);
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.12s;
}
.dropdown-sub-item:hover {
  color: var(--primary, #2d8cf0);
  background: rgba(45, 140, 240, 0.06);
}

.top-nav-right {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.header-action {
  position: relative;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  cursor: pointer;
  color: var(--text-secondary, #666);
  transition: all 0.15s;
}
.header-action:hover {
  background: rgba(0, 0, 0, 0.04);
  color: var(--text-primary, #333);
}

.badge-dot {
  position: absolute;
  top: 4px;
  right: 4px;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  border-radius: 8px;
  background: #f56c6c;
  color: #fff;
  font-size: 10px;
  line-height: 16px;
  text-align: center;
}

.user-info {
  position: relative;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s;
  margin-left: 8px;
}
.user-info:hover {
  background: rgba(0, 0, 0, 0.04);
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
}

.user-name {
  font-size: 14px;
  color: var(--text-primary, #333);
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-dropdown {
  position: absolute;
  top: 100%;
  right: 0;
  min-width: 140px;
  background: var(--card-bg, #fff);
  border: 1px solid var(--border-color, #e8e8e8);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  z-index: 100;
  padding: 4px 0;
  animation: dropdownIn 0.12s ease-out;
}

.user-dropdown-item {
  padding: 10px 16px;
  font-size: 14px;
  color: var(--text-primary, #333);
  cursor: pointer;
  transition: all 0.12s;
}
.user-dropdown-item:hover {
  background: rgba(0, 0, 0, 0.04);
}

.user-dropdown-divider {
  height: 1px;
  background: var(--border-color, #eee);
  margin: 4px 0;
}

.notification-wrapper {
  position: relative;
}

.notification-popup {
  position: absolute;
  top: 100%;
  right: 0;
  width: 320px;
  background: var(--card-bg, #fff);
  border: 1px solid var(--border-color, #e8e8e8);
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
  z-index: 200;
  margin-top: 8px;
  animation: dropdownIn 0.15s ease-out;
}

.notification-popup-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color, #eee);
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #333);
}

.notification-clear {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-secondary, #999);
  cursor: pointer;
}

.notification-clear:hover {
  color: var(--primary, #2d8cf0);
}

.notification-popup-body {
  max-height: 300px;
  overflow-y: auto;
}

.notification-list {
  padding: 4px 0;
}

.notification-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 16px;
  cursor: pointer;
  transition: background 0.12s;
}

.notification-item:hover {
  background: rgba(0, 0, 0, 0.03);
}

.notification-item.unread {
  background: rgba(45, 140, 240, 0.04);
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
  background: var(--primary, #2d8cf0);
}

.notification-content-wrapper {
  flex: 1;
  min-width: 0;
}

.notification-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary, #333);
  margin-bottom: 2px;
}

.notification-content {
  font-size: 12px;
  color: var(--text-secondary, #999);
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notification-time {
  font-size: 11px;
  color: var(--text-tertiary, #bbb);
}

.notification-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px 16px;
  color: var(--text-secondary, #999);
  font-size: 13px;
}
</style>
