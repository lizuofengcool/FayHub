<template>
  <div class="tab-manager" :style="{ height: tabbarHeight + 'px' }">
    <div class="tab-nav-scroll" ref="tabsWrapperRef">
      <div class="tab-nav-list">
        <div
          v-for="tab in tabs"
          :key="tab.id"
          class="tab-item"
          :class="{
            active: tab.active,
            pinned: tab.pinned,
          }"
          @click="switchTab(tab)"
          @contextmenu.prevent="showContextMenu($event, tab)"
        >
          <el-icon class="tab-icon" v-if="tab.icon && iconMap[tab.icon]">
            <component :is="iconMap[tab.icon]" />
          </el-icon>
          <span class="tab-title">{{ tab.title }}</span>
          <span class="tab-close" v-if="tab.closable" @click.stop="closeTab(tab)">
            <svg viewBox="0 0 12 12" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M3 3l6 6M9 3l-6 6"/>
            </svg>
          </span>
        </div>
      </div>
      <div class="tab-bar" :style="barStyle"></div>
    </div>

    <div class="tab-suffix">
      <el-tooltip content="刷新当前页" placement="bottom">
        <button class="suffix-btn" @click="refreshTab">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M1 4v6h6M23 20v-6h-6"/>
            <path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15"/>
          </svg>
        </button>
      </el-tooltip>
      <el-dropdown trigger="click" @command="handleTabCommand">
        <button class="suffix-btn">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
            <path d="m6 9 6 6 6-6"/>
          </svg>
        </button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="close-other">
              <el-icon><Close /></el-icon>关闭其他
            </el-dropdown-item>
            <el-dropdown-item command="close-all">
              <el-icon><CircleClose /></el-icon>关闭所有
            </el-dropdown-item>
            <el-dropdown-item divided command="pin" :disabled="!activeTab">
              <el-icon><Paperclip /></el-icon>{{ activeTab?.pinned ? '取消固定' : '固定标签' }}
            </el-dropdown-item>
            <el-dropdown-item command="refresh">
              <el-icon><RefreshRight /></el-icon>重新加载
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <div
      v-if="contextMenu.visible"
      class="context-menu"
      :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
    >
      <div class="menu-item" @click="contextMenu.tab && refreshTab(contextMenu.tab)">
        <el-icon><RefreshRight /></el-icon>重新加载
      </div>
      <div class="menu-item" @click="contextMenu.tab && closeTab(contextMenu.tab)" v-if="contextMenu.tab?.closable">
        <el-icon><Close /></el-icon>关闭标签
      </div>
      <div class="menu-divider"></div>
      <div class="menu-item" @click="closeOtherTabs(contextMenu.tab)">
        <el-icon><Close /></el-icon>关闭其他
      </div>
      <div class="menu-item" @click="closeLeftTabs(contextMenu.tab)">
        <el-icon><Close /></el-icon>关闭左侧
      </div>
      <div class="menu-item" @click="closeRightTabs(contextMenu.tab)">
        <el-icon><Close /></el-icon>关闭右侧
      </div>
      <div class="menu-item" @click="closeAllTabs">
        <el-icon><CircleClose /></el-icon>关闭全部
      </div>
      <div class="menu-divider"></div>
      <div class="menu-item" @click="pinTab(contextMenu.tab)">
        <el-icon><Paperclip /></el-icon>{{ contextMenu.tab?.pinned ? '取消固定' : '固定标签' }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  Close, RefreshRight, CircleClose, Paperclip,
  Monitor, Setting, User, Lock, Menu, Connection, Shop,
  DataAnalysis, Grid, Key, List, Management, Tickets,
  CreditCard, Wallet, Folder, Upload, Document, Link, Tools, Promotion
} from '@element-plus/icons-vue'
import { usePreferencesStore } from '@/stores/preferences'

interface Tab {
  id: string
  title: string
  path: string
  icon?: string
  active: boolean
  closable: boolean
  pinned: boolean
  query?: Record<string, any>
  params?: Record<string, any>
}

const router = useRouter()
const route = useRoute()
const tabsWrapperRef = ref<HTMLElement>()
const prefsStore = usePreferencesStore()

const tabs = ref<Tab[]>([])
const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  tab: undefined as Tab | undefined
})

const iconMap: Record<string, any> = {
  Monitor, Setting, User, Lock, Menu, Connection, Shop,
  DataAnalysis, Grid, Key, List, Management, Tickets,
  CreditCard, Wallet, Folder, Upload, Document, Link, Tools, Promotion
}

const activeTab = computed(() => tabs.value.find(tab => tab.active))

const tabbarHeight = computed(() => prefsStore.prefs.tabbarHeight)

const barStyle = computed(() => {
  if (!tabsWrapperRef.value) return {}
  const activeEl = tabsWrapperRef.value.querySelector('.tab-item.active') as HTMLElement
  if (!activeEl) return { display: 'none' }
  const scrollLeft = tabsWrapperRef.value.scrollLeft
  const left = activeEl.offsetLeft - scrollLeft
  const width = activeEl.offsetWidth
  return {
    left: left + 'px',
    width: width + 'px',
  }
})

// 生成标签页ID
const generateTabId = (route: any): string => {
  const baseId = route.path
  const queryStr = Object.keys(route.query).sort().map(key => `${key}=${route.query[key]}`).join('&')
  const paramsStr = Object.keys(route.params).sort().map(key => `${key}=${route.params[key]}`).join('&')

  if (queryStr || paramsStr) {
    return `${baseId}?${queryStr}&${paramsStr}`
  }
  return baseId
}

// 根据路由自动管理标签页
watch(() => route.fullPath, () => {
  if (!route || !route.path) return

  try {
    const tabId = generateTabId(route)
    const existingTab = tabs.value.find(tab => tab.id === tabId)

    if (existingTab) {
      tabs.value.forEach(tab => tab.active = false)
      existingTab.active = true
      existingTab.title = (route.meta?.title as string) || '未命名页面'
    } else {
      tabs.value.forEach(tab => tab.active = false)
      const newTab: Tab = {
        id: tabId,
        title: (route.meta?.title as string) || '未命名页面',
        path: route.path,
        icon: route.meta?.icon as string,
        active: true,
        closable: route.path !== '/dashboard',
        pinned: false,
        query: { ...route.query },
        params: { ...route.params }
      }
      tabs.value.push(newTab)
    }

    // 限制标签页数量（最多15个）
    if (tabs.value.length > 15) {
      const inactiveTabs = tabs.value.filter(tab => !tab.active && !tab.pinned)
      if (inactiveTabs.length > 0) {
        const tabToRemove = inactiveTabs[0]
        tabs.value = tabs.value.filter(tab => tab.id !== tabToRemove.id)
      }
    }

    // 滚动到激活的标签
    nextTick(() => {
      scrollToActiveTab()
    })
  } catch (e) {
    console.error('TabManager watch error:', e)
  }
}, { immediate: true })

// 滚动到激活的标签
const scrollToActiveTab = () => {
  if (!tabsWrapperRef.value) return
  const activeEl = tabsWrapperRef.value.querySelector('.tab-item.active') as HTMLElement
  if (activeEl) {
    activeEl.scrollIntoView({ behavior: 'smooth', inline: 'center', block: 'nearest' })
  }
  nextTick(() => {
    updateBarPosition()
  })
}

const updateBarPosition = () => {
  if (!tabsWrapperRef.value) return
  const activeEl = tabsWrapperRef.value.querySelector('.tab-item.active') as HTMLElement
  if (!activeEl) return
  const bar = tabsWrapperRef.value.querySelector('.tab-bar') as HTMLElement
  if (!bar) return
  const scrollLeft = tabsWrapperRef.value.scrollLeft
  bar.style.left = (activeEl.offsetLeft - scrollLeft) + 'px'
  bar.style.width = activeEl.offsetWidth + 'px'
}

// 切换标签页
const switchTab = (tab: Tab) => {
  if (tab.active) return

  tabs.value.forEach(t => t.active = false)
  tab.active = true

  router.push({
    path: tab.path,
    query: tab.query
  })
}

// 刷新标签页
const refreshTab = (tab?: Tab) => {
  const targetTab = tab || activeTab.value
  if (!targetTab) return

  const currentPath = targetTab.path
  router.replace('/redirect' + currentPath).then(() => {
    nextTick(() => {
      router.replace(currentPath)
    })
  })
}

// 关闭标签页
const closeTab = (tab: Tab) => {
  if (!tab.closable) return

  const tabIndex = tabs.value.findIndex(t => t.id === tab.id)
  if (tabIndex === -1) return

  if (tab.active) {
    const remainingTabs = tabs.value.filter(t => t.id !== tab.id)
    if (remainingTabs.length > 0) {
      const nextTab = remainingTabs[Math.max(0, tabIndex - 1)]
      switchTab(nextTab)
    } else {
      router.push('/dashboard')
    }
  }

  tabs.value = tabs.value.filter(t => t.id !== tab.id)
}

// 关闭其他标签页
const closeOtherTabs = (keepTab?: Tab) => {
  const targetTab = keepTab || activeTab.value
  if (!targetTab) return

  tabs.value = tabs.value.filter(tab =>
    tab.id === targetTab.id || tab.pinned || !tab.closable
  )

  tabs.value.forEach(tab => tab.active = tab.id === targetTab.id)
}

// 关闭左侧标签页
const closeLeftTabs = (tab?: Tab) => {
  const targetTab = tab || activeTab.value
  if (!targetTab) return

  const tabIndex = tabs.value.findIndex(t => t.id === targetTab.id)
  if (tabIndex === -1) return

  tabs.value = tabs.value.filter((t, index) =>
    index >= tabIndex || t.pinned || !t.closable
  )
}

// 关闭右侧标签页
const closeRightTabs = (tab?: Tab) => {
  const targetTab = tab || activeTab.value
  if (!targetTab) return

  const tabIndex = tabs.value.findIndex(t => t.id === targetTab.id)
  if (tabIndex === -1) return

  tabs.value = tabs.value.filter((t, index) =>
    index <= tabIndex || t.pinned || !t.closable
  )
}

// 关闭所有标签页
const closeAllTabs = () => {
  const keepTabs = tabs.value.filter(tab => !tab.closable || tab.pinned)

  if (keepTabs.length > 0) {
    tabs.value = keepTabs
    switchTab(keepTabs[0])
  } else {
    tabs.value = []
    router.push('/dashboard')
  }
}

// 固定/取消固定标签页
const pinTab = (tab?: Tab) => {
  const targetTab = tab || activeTab.value
  if (!targetTab) return
  targetTab.pinned = !targetTab.pinned
}

// 显示右键菜单
const showContextMenu = (event: MouseEvent, tab: Tab) => {
  event.preventDefault()
  contextMenu.value = {
    visible: true,
    x: event.clientX,
    y: event.clientY,
    tab
  }

  const closeMenu = () => {
    contextMenu.value.visible = false
    document.removeEventListener('click', closeMenu)
  }
  nextTick(() => {
    document.addEventListener('click', closeMenu)
  })
}

// 处理标签页命令
const handleTabCommand = (command: string) => {
  switch (command) {
    case 'close-other':
      closeOtherTabs()
      break
    case 'close-all':
      closeAllTabs()
      break
    case 'pin':
      pinTab()
      break
    case 'refresh':
      refreshTab()
      break
  }
}

// 键盘快捷键
const handleKeydown = (event: KeyboardEvent) => {
  if (event.ctrlKey || event.metaKey) {
    switch (event.key) {
      case 'w':
      case 'W':
        event.preventDefault()
        if (activeTab.value?.closable) {
          closeTab(activeTab.value)
        }
        break
      case 'r':
      case 'R':
        event.preventDefault()
        refreshTab()
        break
    }
  }
}

document.addEventListener('keydown', handleKeydown)
</script>

<style scoped>
.tab-manager {
  display: flex;
  align-items: center;
  background: var(--header-bg, #fff);
  border-bottom: 1px solid var(--border-color, #e8e8e8);
  padding: 0 8px;
  position: relative;
}

.tab-nav-scroll {
  flex: 1;
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: none;
  position: relative;
  height: 100%;
  display: flex;
  align-items: flex-end;
}
.tab-nav-scroll::-webkit-scrollbar {
  display: none;
}

.tab-nav-list {
  display: flex;
  align-items: flex-end;
  height: 100%;
  gap: 0;
}

.tab-item {
  display: flex;
  align-items: center;
  padding: 0 16px;
  height: calc(100% - 2px);
  cursor: pointer;
  transition: color 0.2s;
  color: var(--text-secondary, #666);
  font-size: 13px;
  position: relative;
  white-space: nowrap;
  user-select: none;
  gap: 6px;
  border-radius: 6px 6px 0 0;
}
.tab-item:hover {
  color: var(--text-primary, #333);
  background: rgba(0, 0, 0, 0.03);
}
.tab-item.active {
  color: var(--primary, #2d8cf0);
  font-weight: 500;
}

.tab-icon {
  font-size: 15px;
  flex-shrink: 0;
}

.tab-title {
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tab-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  color: transparent;
  transition: all 0.15s;
  flex-shrink: 0;
}
.tab-item:hover .tab-close {
  color: var(--text-muted, #999);
}
.tab-close:hover {
  color: #fff !important;
  background: rgba(0, 0, 0, 0.25);
}
.tab-item.active .tab-close {
  color: var(--primary, #2d8cf0);
}
.tab-item.active .tab-close:hover {
  color: #fff !important;
  background: var(--primary, #2d8cf0);
}

.tab-bar {
  position: absolute;
  bottom: 0;
  height: 2px;
  background: var(--primary, #2d8cf0);
  border-radius: 1px 1px 0 0;
  transition: left 0.3s cubic-bezier(0.4, 0, 0.2, 1), width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.tab-suffix {
  display: flex;
  align-items: center;
  gap: 2px;
  padding-left: 8px;
  margin-left: 4px;
  border-left: 1px solid var(--border-color, #e8e8e8);
  flex-shrink: 0;
  height: 100%;
}

.suffix-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border: none;
  background: none;
  color: var(--text-secondary, #666);
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.15s;
}
.suffix-btn:hover {
  color: var(--primary, #2d8cf0);
  background: rgba(45, 140, 240, 0.08);
}

.context-menu {
  position: fixed;
  background: #fff;
  border: 1px solid var(--border-color, #e8e8e8);
  border-radius: 8px;
  box-shadow: 0 6px 24px rgba(0, 0, 0, 0.1);
  z-index: 1000;
  padding: 4px 0;
  min-width: 160px;
  animation: fadeIn 0.15s ease-out;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  font-size: 13px;
  color: var(--text-primary, #333);
  cursor: pointer;
  transition: all 0.1s;
  gap: 8px;
}
.menu-item:hover {
  background: rgba(45, 140, 240, 0.06);
  color: var(--primary, #2d8cf0);
}

.menu-divider {
  height: 1px;
  background: var(--border-color, #e8e8e8);
  margin: 4px 0;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
