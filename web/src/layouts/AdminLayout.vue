<template>
  <div class="admin-layout" :class="[`layout-${prefsStore.prefs.layoutMode}`, { 'sidebar-collapsed': sidebarCollapsed }]">

    <div
      v-if="sidebarOpen && isMobile"
      class="mobile-overlay"
      @click="sidebarOpen = false"
    ></div>

    <!-- ====== 垂直布局 (side) ====== -->
    <template v-if="prefsStore.prefs.layoutMode === 'side'">
      <aside
        class="layout-sidebar"
        :class="{ 'mobile-open': isMobile && sidebarOpen, 'mobile-closed': isMobile && !sidebarOpen }"
        :style="{ width: sidebarCollapsed ? prefsStore.prefs.collapseWidth + 'px' : prefsStore.prefs.sidebarWidth + 'px' }"
      >
        <div class="sidebar-logo" v-if="prefsStore.prefs.logo">
          <div class="logo-icon">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="18" height="18" rx="3"/>
              <path d="M3 9h18M9 21V9"/>
            </svg>
          </div>
          <h1 class="logo-text" v-show="!sidebarCollapsed">FayHub Admin</h1>
        </div>

        <div class="sidebar-menu">
          <ThreeLevelMenu :menus="visibleMenuItems" />
        </div>

        <div class="sidebar-collapse-btn" @click="toggleSidebarCollapse">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" :class="{ rotated: sidebarCollapsed }">
            <path d="m15 18-6-6 6-6"/>
          </svg>
        </div>
      </aside>
    </template>

    <!-- ====== 双栏布局 (mix): 主菜单侧边栏 + 子菜单侧边栏 ====== -->
    <template v-if="prefsStore.prefs.layoutMode === 'mix'">
      <aside
        class="layout-sidebar layout-sidebar-main"
        :class="{ 'mobile-open': isMobile && sidebarOpen, 'mobile-closed': isMobile && !sidebarOpen }"
        :style="{ width: '72px' }"
      >
        <div class="sidebar-logo sidebar-logo-compact" v-if="prefsStore.prefs.logo">
          <div class="logo-icon logo-icon-sm">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="18" height="18" rx="3"/>
              <path d="M3 9h18M9 21V9"/>
            </svg>
          </div>
        </div>
        <div class="sidebar-menu">
          <div
            v-for="menu in visibleMenuItems"
            :key="menu.id"
            class="main-menu-icon-item"
            :class="{ active: activeMainMenuId === menu.id }"
            @click="selectMainMenu(menu)"
            :title="menu.title"
          >
            <span class="main-menu-icon-dot"></span>
            <span class="main-menu-label">{{ menu.title }}</span>
          </div>
        </div>
      </aside>

      <aside
        v-if="activeMainMenuChildren.length > 0"
        class="layout-sidebar layout-sidebar-sub"
        :style="{ width: prefsStore.prefs.sidebarWidth + 'px' }"
      >
        <div class="sub-sidebar-header">
          <span>{{ activeMainMenuTitle }}</span>
        </div>
        <div class="sidebar-menu">
          <div
            v-for="child in activeMainMenuChildren"
            :key="child.id"
            class="sub-menu-item"
            :class="{ active: route.path === child.path || route.path.startsWith(child.path + '/') }"
            @click="router.push(child.path)"
          >
            {{ child.title }}
          </div>
        </div>
      </aside>
    </template>

    <!-- ====== 水平布局 (top): 顶部导航 ====== -->
    <template v-if="prefsStore.prefs.layoutMode === 'top'">
      <TopNavHeader
        :menu-items="visibleMenuItems"
        :user-info="userInfo"
        :unread-count="unreadCount"
        :current-page-title="currentPageTitle"
        @open-search="showSearch = true"
        @open-settings="showSettings = true"
        @logout="handleLogout"
      />
    </template>

    <!-- ====== 混合双栏 (mix-double): 顶部导航 + 双侧边栏 ====== -->
    <template v-if="prefsStore.prefs.layoutMode === 'mix-double'">
      <TopNavHeader
        mode="click"
        :menu-items="visibleMenuItems"
        :user-info="userInfo"
        :unread-count="unreadCount"
        :current-page-title="currentPageTitle"
        @open-search="showSearch = true"
        @open-settings="showSettings = true"
        @logout="handleLogout"
        @menu-select="handleTopMenuSelect"
      />
      <aside class="layout-sub-sidebar layout-sub-sidebar-main">
        <div class="sub-sidebar-menu">
          <div
            v-for="child in activeTopMenuChildren"
            :key="child.id"
            class="sub-menu-item sub-menu-item-main"
            :class="{ active: activeSubMenuId === child.id }"
            @click="selectSubMenu(child)"
          >
            {{ child.title }}
          </div>
        </div>
      </aside>
      <aside class="layout-sub-sidebar layout-sub-sidebar-sub">
        <div class="sub-sidebar-header">
          <span>{{ activeSubMenuTitle }}</span>
        </div>
        <div class="sub-sidebar-menu">
          <div
            v-for="grandchild in activeSubMenuChildren"
            :key="grandchild.id"
            class="sub-menu-item"
            :class="{ active: route.path === grandchild.path || route.path.startsWith(grandchild.path + '/') }"
            @click="router.push(grandchild.path)"
          >
            {{ grandchild.title }}
          </div>
        </div>
      </aside>
    </template>

    <!-- ====== 混合侧边栏 (mix-sidebar): 顶部导航 + 单侧边栏 ====== -->
    <template v-if="prefsStore.prefs.layoutMode === 'mix-sidebar'">
      <TopNavHeader
        mode="click"
        :menu-items="visibleMenuItems"
        :user-info="userInfo"
        :unread-count="unreadCount"
        :current-page-title="currentPageTitle"
        @open-search="showSearch = true"
        @open-settings="showSettings = true"
        @logout="handleLogout"
        @menu-select="navigateTopMenu"
      />
      <aside class="layout-sub-sidebar" :style="{ width: prefsStore.prefs.sidebarWidth + 'px' }">
        <div class="sub-sidebar-header">
          <span>{{ activeTopMenuTitle }}</span>
        </div>
        <div class="sub-sidebar-menu">
          <div
            v-for="child in activeTopMenuChildren"
            :key="child.id"
            class="sub-menu-item"
            :class="{ active: route.path === child.path || route.path.startsWith(child.path + '/') }"
            @click="router.push(child.path)"
          >
            {{ child.title }}
          </div>
        </div>
      </aside>
    </template>

    <!-- ====== 主内容区 ====== -->
    <div class="layout-main">
      <EnhancedHeader
        v-if="prefsStore.prefs.headerVisible && showInnerHeader"
        :user-info="userInfo"
        :unread-count="unreadCount"
        :current-page-title="currentPageTitle"
        @logout="handleLogout"
        @open-settings="showSettings = true"
        @open-search="showSearch = true"
      />

      <div v-if="isImpersonated" class="impersonate-bar">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/></svg>
        <span>您正在以超级管理员身份模拟登录租户后台，所有操作将以租户管理员身份执行</span>
        <button class="impersonate-exit" @click="exitImpersonate">退出模拟</button>
      </div>

      <TabManager v-if="prefsStore.prefs.tabbarVisible" />

      <div class="layout-content" :class="{ 'content-gap': prefsStore.prefs.gap }">
        <router-view />
      </div>

      <button
        v-if="prefsStore.prefs.layoutMode === 'full'"
        class="fullscreen-exit-btn"
        title="退出全屏"
        @click="prefsStore.setLayoutMode(prefsStore.prefs.defaultLayout || 'side')"
      >
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M8 3v3a2 2 0 0 1-2 2H3m18 0h-3a2 2 0 0 1-2-2V3m0 18v-3a2 2 0 0 1 2-2h3M3 16h3a2 2 0 0 1 2 2v3"/>
        </svg>
      </button>

      <div class="layout-footer" v-if="prefsStore.prefs.footerVisible">
        <span>FayHub Admin &copy; 2026 &middot; Made with &#10084;&#65039;</span>
      </div>
    </div>

    <PluginDevTools v-if="showDevTools" />
    <SettingsPanel :visible="showSettings" @close="showSettings = false" />
    <SearchDialog :visible="showSearch" @close="showSearch = false" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMessage, useDialog } from 'naive-ui'
import menuApi, { type Menu as MenuType } from '@/api/menu'
import notificationApi from '@/api/notification'
import { useUserStore } from '@/stores/user'

import PluginDevTools from '@/plugin/PluginDevTools.vue'
import ThreeLevelMenu from './components/ThreeLevelMenu.vue'
import TabManager from './components/TabManager.vue'
import EnhancedHeader from './components/EnhancedHeader.vue'
import TopNavHeader from './components/TopNavHeader.vue'
import SettingsPanel from '@/components/SettingsPanel.vue'
import SearchDialog from '@/components/SearchDialog.vue'
import { usePreferencesStore } from '@/stores/preferences'

const showDevTools = import.meta.env.DEV

const router = useRouter()
const route = useRoute()

const prefsStore = usePreferencesStore()

const message = useMessage()
const dialog = useDialog()

const isMobile = ref(false)
const sidebarOpen = ref(false)
const sidebarCollapsed = ref(false)
const showSettings = ref(false)
const showSearch = ref(false)
const activeTopMenuId = ref<number | null>(null)
const activeMainMenuId = ref<number | null>(null)
const activeSubMenuId = ref<number | null>(null)

// 根据当前路由初始化选中的菜单
function initActiveMenu() {
  const currentPath = route.path
  for (const menu of menuItems.value) {
    if (menu.children) {
      for (const child of menu.children) {
        if (child.path === currentPath || currentPath.startsWith(child.path + '/')) {
          activeTopMenuId.value = menu.id
          activeMainMenuId.value = menu.id
          activeSubMenuId.value = child.id
          return
        }
      }
    }
    if (menu.path === currentPath) {
      activeTopMenuId.value = menu.id
      activeMainMenuId.value = menu.id
      return
    }
  }
}

function checkMobile() {
  isMobile.value = window.innerWidth < 768
  if (!isMobile.value) sidebarOpen.value = false
}

function toggleSidebarCollapse() {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

function handleGlobalKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault()
    showSearch.value = !showSearch.value
  }
  if ((e.ctrlKey || e.metaKey) && e.key === 'p') {
    e.preventDefault()
    showSettings.value = !showSettings.value
  }
}

const registeredPluginRoutes = ref<Set<string>>(new Set())

const userInfo = ref({
  id: 1,
  username: 'admin',
  role: '系统管理员',
  avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=admin&backgroundColor=e2e8f0'
})

const menuItems = ref<MenuType[]>([])

const unreadCount = ref(0)

async function fetchUnreadCount() {
  try {
    const res = await notificationApi.getUnreadCount()
    unreadCount.value = res.data?.unread_count || 0
  } catch {}
}

const visibleMenuItems = computed(() => {
  return menuItems.value.filter(menu => {
    if (menu.children && menu.children.length > 0) {
      const activeChildren = menu.children.filter((child: MenuType) => child.status !== 0)
      if (activeChildren.length === 0) return false
    }
    if (menu.path === '/plugin-apps' && (!menu.children || menu.children.filter((c: MenuType) => c.status !== 0).length === 0)) {
      return false
    }
    return true
  }).map(menu => {
    if (menu.children && menu.children.length > 0) {
      return { ...menu, children: menu.children.filter((child: MenuType) => child.status !== 0) }
    }
    return menu
  })
})

const isImpersonated = computed(() => {
  return !!localStorage.getItem('fayhub_impersonated_tenant')
})

function exitImpersonate() {
  localStorage.removeItem('fayhub_token')
  localStorage.removeItem('fayhub_refresh_token')
  localStorage.removeItem('fayhub_impersonated_tenant')
  localStorage.removeItem('userInfo')
  message.success('已退出模拟登录')
  window.location.replace('/')
}

const currentPageTitle = computed(() => {
  const routeMeta = route.meta as { title?: string }
  return routeMeta.title || '系统管理'
})

const showInnerHeader = computed(() => {
  const mode = prefsStore.prefs.layoutMode
  return mode !== 'top' && mode !== 'mix-double' && mode !== 'mix-sidebar' && mode !== 'full'
})

function isMenuActive(menu: MenuType): boolean {
  if (activeTopMenuId.value === menu.id) return true
  if (menu.children) {
    return menu.children.some((child: MenuType) =>
      route.path === child.path || route.path.startsWith(child.path + '/')
    )
  }
  return route.path === menu.path
}

const activeTopMenuTitle = computed(() => {
  if (activeTopMenuId.value) {
    const menu = menuItems.value.find(m => m.id === activeTopMenuId.value)
    return menu?.title || ''
  }
  const activeMenu = menuItems.value.find(m => isMenuActive(m))
  return activeMenu?.title || ''
})

const activeTopMenuChildren = computed(() => {
  if (!activeTopMenuId.value) {
    const activeMenu = menuItems.value.find(m => isMenuActive(m))
    return activeMenu?.children || []
  }
  const menu = menuItems.value.find(m => m.id === activeTopMenuId.value)
  return menu?.children || []
})

function navigateTopMenu(menu: MenuType) {
  activeTopMenuId.value = menu.id
  if (menu.children && menu.children.length > 0) {
    const firstChild = menu.children[0]
    if (firstChild.path) router.push(firstChild.path)
  } else if (menu.path) {
    router.push(menu.path)
  }
}

function handleTopMenuSelect(menu: { id: number; title: string; path?: string; children?: any[] }) {
  activeTopMenuId.value = menu.id
  activeSubMenuId.value = null
  if (menu.children && menu.children.length > 0) {
    const firstChild = menu.children[0]
    if (firstChild.children && firstChild.children.length > 0) {
      activeSubMenuId.value = firstChild.id
      const firstGrandchild = firstChild.children[0]
      if (firstGrandchild.path) router.push(firstGrandchild.path)
    } else if (firstChild.path) {
      router.push(firstChild.path)
    }
  } else if (menu.path) {
    router.push(menu.path)
  }
}

// 双栏布局 (mix) 主菜单选择
function selectMainMenu(menu: MenuType) {
  activeMainMenuId.value = menu.id
  activeSubMenuId.value = null
  if (menu.children && menu.children.length > 0) {
    const firstChild = menu.children[0]
    if (firstChild.path) router.push(firstChild.path)
  } else if (menu.path) {
    router.push(menu.path)
  }
}

const activeMainMenuTitle = computed(() => {
  if (activeMainMenuId.value) {
    const menu = menuItems.value.find(m => m.id === activeMainMenuId.value)
    return menu?.title || ''
  }
  const activeMenu = menuItems.value.find(m => isMenuActive(m))
  return activeMenu?.title || ''
})

const activeMainMenuChildren = computed(() => {
  if (!activeMainMenuId.value) {
    const activeMenu = menuItems.value.find(m => isMenuActive(m))
    return activeMenu?.children || []
  }
  const menu = menuItems.value.find(m => m.id === activeMainMenuId.value)
  return menu?.children || []
})

// 混合双栏 (mix-double) 子菜单选择
function selectSubMenu(child: MenuType) {
  activeSubMenuId.value = child.id
  if (child.children && child.children.length > 0) {
    const firstGrandchild = child.children[0]
    if (firstGrandchild.path) router.push(firstGrandchild.path)
  } else if (child.path) {
    router.push(child.path)
  }
}

const activeSubMenuTitle = computed(() => {
  if (activeSubMenuId.value) {
    const children = activeTopMenuChildren.value
    const child = children.find(c => c.id === activeSubMenuId.value)
    return child?.title || ''
  }
  return ''
})

const activeSubMenuChildren = computed(() => {
  if (activeSubMenuId.value) {
    const children = activeTopMenuChildren.value
    const child = children.find(c => c.id === activeSubMenuId.value)
    return child?.children || []
  }
  return []
})

const handleLogout = () => {
  dialog.warning({
    title: '提示',
    content: '确定要退出登录吗？',
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      const userStore = useUserStore()
      await userStore.logout()
      message.success('已安全退出')
      window.location.replace('/')
    }
  })
}

async function fetchMenus() {
  try {
    const res = await menuApi.getMenuTree()
    menuItems.value = res.data || []
    registerPluginRoutes(menuItems.value)
    initActiveMenu()
  } catch (err: any) {
    console.error('获取菜单失败:', err)
  }
}

function registerPluginRoutes(menus: MenuType[]) {
  for (const menu of menus) {
    if (menu.children && menu.children.length > 0) {
      for (const child of menu.children) {
        if (child.path && child.component && !registeredPluginRoutes.value.has(child.path)) {
          const existingRoute = router.getRoutes().find(r => r.path === child.path)
          if (!existingRoute) {
            router.addRoute('admin', {
              path: child.path,
              name: `plugin-${child.path.replace(/\//g, '-')}`,
              component: () => import('@/views/PluginPage.vue'),
              meta: { requiresAuth: true, title: child.title, pluginId: child.component },
              props: { pluginId: child.component }
            })
          }
          registeredPluginRoutes.value.add(child.path)
        }
      }
    }
  }
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  prefsStore.applyAll()
  window.addEventListener('keydown', handleGlobalKeydown)
  window.addEventListener('menu-refresh', fetchMenus)

  const storedUser = localStorage.getItem('userInfo')
  if (storedUser) {
    try {
      const parsed = JSON.parse(storedUser)
      userInfo.value = { ...userInfo.value, ...parsed }
    } catch {}
  }

  fetchMenus()
  fetchUnreadCount()
})

watch(() => route.path, () => {
  const refreshFlag = localStorage.getItem('menu_refresh_needed')
  if (refreshFlag === 'true') {
    localStorage.removeItem('menu_refresh_needed')
    fetchMenus()
  }
  initActiveMenu()
})

let menuRefreshTimer: ReturnType<typeof setInterval> | null = null
onMounted(() => {
  menuRefreshTimer = setInterval(() => {
    const refreshFlag = localStorage.getItem('menu_refresh_needed')
    if (refreshFlag === 'true') {
      localStorage.removeItem('menu_refresh_needed')
      fetchMenus()
    }
    fetchUnreadCount()
  }, 30000)
})
onBeforeUnmount(() => {
  if (menuRefreshTimer) {
    clearInterval(menuRefreshTimer)
    menuRefreshTimer = null
  }
  window.removeEventListener('resize', checkMobile)
  window.removeEventListener('keydown', handleGlobalKeydown)
  window.removeEventListener('menu-refresh', fetchMenus)
})
</script>

<style scoped>
.admin-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  background: var(--body-bg, #f0f2f5);
  color: var(--text-primary, #333);
}

.mobile-overlay {
  display: none;
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  z-index: 30;
}
@media (max-width: 767px) {
  .mobile-overlay { display: block; }
}

/* ====== 侧边栏通用 ====== */
.layout-sidebar {
  display: flex;
  flex-direction: column;
  background: var(--sidebar-bg, #001529);
  color: var(--sidebar-text, rgba(255,255,255,0.85));
  transition: width 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  flex-shrink: 0;
  overflow: hidden;
  z-index: 40;
  border-right: 1px solid var(--sidebar-border, rgba(255, 255, 255, 0.08));
}
@media (max-width: 767px) {
  .layout-sidebar {
    position: fixed;
    inset: 0 auto 0 0;
    width: 240px !important;
  }
  .layout-sidebar.mobile-closed {
    transform: translateX(-100%);
  }
  .layout-sidebar.mobile-open {
    transform: translateX(0);
  }
}

.sidebar-logo {
  display: flex;
  align-items: center;
  height: var(--nav-height, 56px);
  padding: 0 20px;
  border-bottom: 1px solid var(--sidebar-border, rgba(255, 255, 255, 0.08));
  flex-shrink: 0;
  gap: 10px;
}
.sidebar-logo-compact {
  justify-content: center;
  padding: 0;
}
.logo-icon {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  background: var(--primary, #2d8cf0);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}
.logo-icon-sm {
  width: 30px;
  height: 30px;
  border-radius: 6px;
}
.logo-text {
  font-size: 17px;
  font-weight: 700;
  color: var(--text-title, #333);
  white-space: nowrap;
  overflow: hidden;
}

.sidebar-menu {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 8px 0;
}
.sidebar-menu::-webkit-scrollbar { width: 4px; }
.sidebar-menu::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.15); border-radius: 2px; }

.sidebar-collapse-btn {
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-top: 1px solid var(--sidebar-border, rgba(255, 255, 255, 0.08));
  cursor: pointer;
  color: var(--sidebar-text, rgba(255, 255, 255, 0.5));
  transition: all 0.15s;
  flex-shrink: 0;
}
.sidebar-collapse-btn:hover {
  color: var(--sidebar-text-active, rgba(255, 255, 255, 0.85));
  background: var(--sidebar-hover, rgba(255, 255, 255, 0.04));
}
.sidebar-collapse-btn svg {
  transition: transform 0.25s;
}
.sidebar-collapse-btn svg.rotated {
  transform: rotate(180deg);
}

/* ====== 双栏布局 (mix) 主菜单侧边栏 ====== */
.layout-sidebar-main {
  width: 72px !important;
  align-items: center;
  border-right: 1px solid var(--sidebar-border, rgba(255, 255, 255, 0.08));
}
.main-menu-icon-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 12px 4px;
  cursor: pointer;
  transition: all 0.15s;
  gap: 4px;
  position: relative;
}
.main-menu-icon-item:hover {
  background: var(--sidebar-hover, rgba(255, 255, 255, 0.06));
}
.main-menu-icon-item.active {
  background: var(--primary-suppl, rgba(45, 140, 240, 0.08));
}
.main-menu-icon-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 8px;
  bottom: 8px;
  width: 3px;
  background: var(--primary, #2d8cf0);
  border-radius: 0 2px 2px 0;
}
.main-menu-icon-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--sidebar-text, rgba(255, 255, 255, 0.4));
  flex-shrink: 0;
}
.main-menu-icon-item.active .main-menu-icon-dot {
  background: var(--primary, #2d8cf0);
}
.main-menu-label {
  font-size: 12px;
  color: var(--sidebar-text, rgba(255, 255, 255, 0.55));
  text-align: center;
  line-height: 1.3;
  max-width: 64px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.main-menu-icon-item.active .main-menu-label {
  color: var(--primary, #2d8cf0);
}

/* ====== 双栏布局 (mix) 子菜单侧边栏 ====== */
.layout-sidebar-sub {
  background: var(--sidebar-sub-bg, #001c3a);
  border-right: 1px solid var(--sidebar-border, rgba(255, 255, 255, 0.08));
}
.sub-sidebar-header {
  display: flex;
  align-items: center;
  height: var(--nav-height, 56px);
  padding: 0 20px;
  border-bottom: 1px solid var(--sidebar-border, rgba(255, 255, 255, 0.06));
  flex-shrink: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--sidebar-text-active, rgba(255, 255, 255, 0.85));
}
.sub-menu-item {
  padding: 10px 20px;
  cursor: pointer;
  font-size: 14px;
  color: var(--sidebar-text, rgba(255, 255, 255, 0.65));
  transition: all 0.15s;
}
.sub-menu-item:hover {
  color: var(--sidebar-text-active, rgba(255, 255, 255, 0.85));
  background: var(--sidebar-hover, rgba(255, 255, 255, 0.04));
}
.sub-menu-item.active {
  color: var(--primary, #2d8cf0);
  background: var(--primary-suppl, rgba(45, 140, 240, 0.08));
  font-weight: 500;
}

/* ====== 混合双栏 (mix-double) / 混合侧边栏 (mix-sidebar) 子侧边栏 ====== */
.layout-sub-sidebar {
  position: absolute;
  top: var(--nav-height, 56px);
  bottom: 0;
  background: var(--sidebar-bg, #fff);
  border-right: 1px solid var(--border-color, #e8e8e8);
  z-index: 20;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}
.layout-sub-sidebar-main {
  left: 0;
  width: 160px;
}
.layout-sub-sidebar-sub {
  left: 160px;
  width: 180px;
  background: var(--sidebar-sub-bg, #fafafa);
}
.layout-mix-double .layout-sub-sidebar .sub-sidebar-header {
  color: var(--text-primary, #333);
  border-bottom-color: var(--border-color, #e8e8e8);
}
.layout-mix-double .layout-sub-sidebar .sub-menu-item {
  color: var(--text-secondary, #666);
}
.layout-mix-double .layout-sub-sidebar .sub-menu-item:hover {
  color: var(--text-primary, #333);
  background: rgba(0, 0, 0, 0.04);
}
.layout-mix-double .layout-sub-sidebar .sub-menu-item.active {
  color: var(--primary, #2d8cf0);
  background: var(--primary-suppl, rgba(45, 140, 240, 0.08));
}

.sub-sidebar-menu {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}
.sub-sidebar-menu::-webkit-scrollbar { width: 4px; }
.sub-sidebar-menu::-webkit-scrollbar-thumb { background: rgba(0,0,0,0.1); border-radius: 2px; }

.sub-menu-item-main {
  font-weight: 500;
  font-size: 14px;
  padding: 10px 16px;
}

/* ====== 主内容区 ====== */
.layout-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

/* 水平布局：顶部留出导航高度 */
.layout-top .layout-main {
  padding-top: var(--nav-height, 56px);
}

/* 混合双栏：左侧留出双侧边栏宽度 */
.layout-mix-double {
  flex-direction: column;
}
.layout-mix-double .layout-main {
  margin-left: 340px;
  padding-top: var(--nav-height, 56px);
}

/* 混合侧边栏：左侧留出侧边栏宽度 */
.layout-mix-sidebar {
  flex-direction: column;
}
.layout-mix-sidebar .layout-main {
  margin-left: v-bind('prefsStore.prefs.sidebarWidth + "px"');
  padding-top: var(--nav-height, 56px);
}

/* 内容全屏 */
.layout-full .layout-main {
  padding: 0;
}

.impersonate-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: #faad14;
  color: #fff;
  font-size: 13px;
  font-weight: 500;
  flex-shrink: 0;
}
.impersonate-exit {
  margin-left: auto;
  padding: 4px 12px;
  border: 1px solid rgba(255,255,255,0.5);
  border-radius: 4px;
  background: transparent;
  color: #fff;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}
.impersonate-exit:hover {
  background: rgba(255,255,255,0.15);
}

.layout-content {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 0;
}
.layout-content::-webkit-scrollbar { width: 6px; }
.layout-content::-webkit-scrollbar-thumb { background: rgba(0,0,0,0.12); border-radius: 3px; }

.content-gap {
  margin: 0;
  border-radius: 0;
  background: var(--body-bg, #f5f6f8);
  box-shadow: none;
  padding: 12px 16px;
}

.layout-footer {
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-top: 1px solid var(--border-color, #e8e8e8);
  font-size: 12px;
  color: var(--text-muted, #999);
  flex-shrink: 0;
  background: var(--card-bg, #fff);
}

.fullscreen-exit-btn {
  position: fixed;
  top: 24px;
  right: 24px;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--card-bg, #fff);
  border: 1px solid var(--border-color, #e8e8e8);
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 100;
  color: var(--text-secondary, #666);
  transition: all 0.2s;
}
.fullscreen-exit-btn:hover {
  color: var(--primary, #2d8cf0);
  border-color: var(--primary, #2d8cf0);
  box-shadow: 0 4px 16px rgba(45, 140, 240, 0.25);
  transform: scale(1.05);
}
</style>
