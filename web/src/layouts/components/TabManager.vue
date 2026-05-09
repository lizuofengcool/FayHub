<template>
  <div class="tab-manager" :style="{ height: tabbarHeight + 'px' }">
    <div class="tab-nav-scroll" ref="tabsWrapperRef">
      <div class="tab-nav-list">
        <div
          v-for="(tab, index) in tabs"
          :key="tab.id"
          class="tab-item"
          :class="{
            active: tab.active,
            pinned: tab.pinned,
            'drag-over': dragOverIndex === index && dragIndex !== index,
          }"
          draggable="true"
          @click="switchTab(tab)"
          @contextmenu.prevent="showContextMenu($event, tab)"
          @dragstart="handleDragStart($event, index)"
          @dragover.prevent="handleDragOver($event, index)"
          @dragleave="handleDragLeave"
          @drop="handleDrop($event, index)"
          @dragend="handleDragEnd"
        >
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
      <n-tooltip trigger="hover">
        <template #trigger>
          <button class="suffix-btn" @click="toggleFullscreen">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/>
            </svg>
          </button>
        </template>
        内容全屏
      </n-tooltip>
      <n-tooltip trigger="hover">
        <template #trigger>
          <button class="suffix-btn" @click="refreshTab()">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M1 4v6h6M23 20v-6h-6"/>
              <path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15"/>
            </svg>
          </button>
        </template>
        刷新当前页
      </n-tooltip>
      <n-dropdown trigger="click" :options="dropdownOptions" @select="handleTabCommand">
        <button class="suffix-btn">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="5" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="12" cy="19" r="1"/>
          </svg>
        </button>
      </n-dropdown>
    </div>

    <div
      v-if="contextMenu.visible"
      class="context-menu"
      :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
    >
      <div class="menu-item" @click="contextMenu.tab && refreshTab(contextMenu.tab)">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 4v6h6M23 20v-6h-6"/><path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15"/></svg>
        重新加载
      </div>
      <div class="menu-item" @click="contextMenu.tab && closeTab(contextMenu.tab)" v-if="contextMenu.tab?.closable">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 6L6 18M6 6l12 12"/></svg>
        关闭标签
      </div>
      <div class="menu-divider"></div>
      <div class="menu-item" @click="closeOtherTabs(contextMenu.tab)">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 6L6 18M6 6l12 12"/></svg>
        关闭其他
      </div>
      <div class="menu-item" @click="closeLeftTabs(contextMenu.tab)">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 6L6 18M6 6l12 12"/></svg>
        关闭左侧
      </div>
      <div class="menu-item" @click="closeRightTabs(contextMenu.tab)">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 6L6 18M6 6l12 12"/></svg>
        关闭右侧
      </div>
      <div class="menu-item" @click="closeAllTabs">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M15 9l-6 6M9 9l6 6"/></svg>
        关闭全部
      </div>
      <div class="menu-divider"></div>
      <div class="menu-item" @click="pinTab(contextMenu.tab)">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>
        {{ contextMenu.tab?.pinned ? '取消固定' : '固定标签' }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, h, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NIcon, NTooltip, NDropdown } from 'naive-ui'
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

const TABS_STORAGE_KEY = 'fayhub_tabs'

function loadTabs(): Tab[] {
  try {
    const saved = localStorage.getItem(TABS_STORAGE_KEY)
    if (saved) {
      return JSON.parse(saved)
    }
  } catch {}
  return []
}

const tabs = ref<Tab[]>(loadTabs())
const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  tab: undefined as Tab | undefined
})

const dragIndex = ref(-1)
const dragOverIndex = ref(-1)
const scrollLeft = ref(0)

function saveTabs() {
  try {
    const data = tabs.value.map(t => ({
      id: t.id,
      title: t.title,
      path: t.path,
      icon: t.icon,
      active: t.active,
      closable: t.closable,
      pinned: t.pinned,
      query: t.query,
      params: t.params
    }))
    localStorage.setItem(TABS_STORAGE_KEY, JSON.stringify(data))
  } catch {}
}

const renderIcon = (iconName: string) => {
  return () => h(NIcon, null, {
    default: () => h('i', { class: `ri-${iconName}` })
  })
}

const dropdownOptions = computed(() => [
  {
    label: '关闭当前',
    key: 'close-current',
    icon: renderIcon('close-line')
  },
  {
    label: '关闭其他',
    key: 'close-other',
    icon: renderIcon('close-line')
  },
  {
    label: '关闭左侧',
    key: 'close-left',
    icon: renderIcon('close-line')
  },
  {
    label: '关闭右侧',
    key: 'close-right',
    icon: renderIcon('close-line')
  },
  {
    label: '关闭所有',
    key: 'close-all',
    icon: renderIcon('close-circle-line')
  },
  {
    type: 'divider' as const,
    key: 'd1'
  },
  {
    label: activeTab.value?.pinned ? '取消固定' : '固定标签',
    key: 'pin',
    disabled: !activeTab.value,
    icon: renderIcon('pushpin-line')
  },
  {
    type: 'divider' as const,
    key: 'd2'
  },
  {
    label: '重新加载',
    key: 'refresh',
    icon: renderIcon('refresh-line')
  }
])

const activeTab = computed(() => tabs.value.find(tab => tab.active))

const tabbarHeight = computed(() => prefsStore.prefs.tabbarHeight)

const barStyle = computed(() => {
  if (!tabsWrapperRef.value) return {}
  const activeEl = tabsWrapperRef.value.querySelector('.tab-item.active') as HTMLElement
  if (!activeEl) return { display: 'none' }
  const left = activeEl.offsetLeft - scrollLeft.value
  const width = activeEl.offsetWidth
  return {
    left: left + 'px',
    width: width + 'px',
  }
})

const generateTabId = (route: any): string => {
  const baseId = route.path
  const queryStr = Object.keys(route.query).sort().map(key => `${key}=${route.query[key]}`).join('&')
  const paramsStr = Object.keys(route.params).sort().map(key => `${key}=${route.params[key]}`).join('&')

  if (queryStr || paramsStr) {
    return `${baseId}?${queryStr}&${paramsStr}`
  }
  return baseId
}

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

    if (tabs.value.length > 15) {
      const inactiveTabs = tabs.value.filter(tab => !tab.active && !tab.pinned)
      if (inactiveTabs.length > 0) {
        const tabToRemove = inactiveTabs[0]
        tabs.value = tabs.value.filter(tab => tab.id !== tabToRemove.id)
      }
    }

    nextTick(() => {
      scrollToActiveTab()
    })
    saveTabs()
  } catch (e) {
    console.error('TabManager watch error:', e)
  }
}, { immediate: true })

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
  scrollLeft.value = tabsWrapperRef.value.scrollLeft
}

const switchTab = (tab: Tab) => {
  if (tab.active) return

  tabs.value.forEach(t => t.active = false)
  tab.active = true

  router.push({
    path: tab.path,
    query: tab.query
  })
}

const refreshTab = (tab?: Tab) => {
  const targetTab = tab || activeTab.value
  if (!targetTab) return

  const currentPath = targetTab.path || route.path
  if (!currentPath) return

  router.replace('/redirect' + currentPath).then(() => {
    nextTick(() => {
      router.replace(currentPath)
    })
  })
}

const closeTab = (tab: Tab) => {
  if (!tab.closable) return

  const tabIndex = tabs.value.findIndex(t => t.id === tab.id)
  if (tabIndex === -1) return

  if (tab.active) {
    const remainingTabs = tabs.value.filter(t => t.id !== tab.id)
    if (remainingTabs.length > 0) {
      const nextTab = remainingTabs[Math.min(tabIndex, remainingTabs.length - 1)]
      switchTab(nextTab)
    }
  }

  tabs.value = tabs.value.filter(t => t.id !== tab.id)
  saveTabs()
}

const closeOtherTabs = (tab?: Tab) => {
  const targetTab = tab || activeTab.value
  if (!targetTab) return

  tabs.value = tabs.value.filter(t => t.id === targetTab.id || t.pinned)
  if (!targetTab.active) {
    switchTab(targetTab)
  }
  saveTabs()
}

const closeLeftTabs = (tab?: Tab) => {
  const targetTab = tab || activeTab.value
  if (!targetTab) return

  const tabIndex = tabs.value.findIndex(t => t.id === targetTab.id)
  if (tabIndex === -1) return

  tabs.value = tabs.value.filter((t, i) => i >= tabIndex || t.pinned)
  saveTabs()
}

const closeRightTabs = (tab?: Tab) => {
  const targetTab = tab || activeTab.value
  if (!targetTab) return

  const tabIndex = tabs.value.findIndex(t => t.id === targetTab.id)
  if (tabIndex === -1) return

  tabs.value = tabs.value.filter((t, i) => i <= tabIndex || t.pinned)
  saveTabs()
}

const closeAllTabs = () => {
  tabs.value = tabs.value.filter(t => t.pinned)
  if (tabs.value.length > 0) {
    switchTab(tabs.value[0])
  }
  saveTabs()
}

const pinTab = (tab?: Tab) => {
  const targetTab = tab || activeTab.value
  if (!targetTab) return

  targetTab.pinned = !targetTab.pinned
  saveTabs()
}

const handleTabCommand = (key: string) => {
  switch (key) {
    case 'close-current':
      if (activeTab.value) closeTab(activeTab.value)
      break
    case 'close-other':
      closeOtherTabs()
      break
    case 'close-left':
      closeLeftTabs()
      break
    case 'close-right':
      closeRightTabs()
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

const showContextMenu = (e: MouseEvent, tab: Tab) => {
  contextMenu.value = {
    visible: true,
    x: e.clientX,
    y: e.clientY,
    tab
  }
}

const hideContextMenu = () => {
  contextMenu.value.visible = false
}

const handleDragStart = (e: DragEvent, index: number) => {
  dragIndex.value = index
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
  }
}

const handleDragOver = (e: DragEvent, index: number) => {
  dragOverIndex.value = index
}

const handleDragLeave = () => {
  dragOverIndex.value = -1
}

const handleDrop = (_e: DragEvent, index: number) => {
  if (dragIndex.value === -1 || dragIndex.value === index) return

  const draggedTab = tabs.value[dragIndex.value]
  tabs.value.splice(dragIndex.value, 1)
  tabs.value.splice(index, 0, draggedTab)

  dragIndex.value = -1
  dragOverIndex.value = -1
  saveTabs()
}

const handleDragEnd = () => {
  dragIndex.value = -1
  dragOverIndex.value = -1
}

const toggleFullscreen = () => {
  prefsStore.setLayoutMode('full')
}

function handleScroll() {
  if (tabsWrapperRef.value) {
    scrollLeft.value = tabsWrapperRef.value.scrollLeft
  }
}

onMounted(() => {
  if (tabsWrapperRef.value) {
    tabsWrapperRef.value.addEventListener('scroll', handleScroll, { passive: true })
  }
})

onUnmounted(() => {
  if (tabsWrapperRef.value) {
    tabsWrapperRef.value.removeEventListener('scroll', handleScroll)
  }
})

document.addEventListener('click', hideContextMenu)
</script>

<style scoped>
.tab-manager {
  display: flex;
  align-items: center;
  background: var(--card-bg, #fff);
  border-bottom: 1px solid var(--border-color, #e8e8e8);
  flex-shrink: 0;
  position: relative;
}

.tab-nav-scroll {
  flex: 1;
  overflow-x: auto;
  overflow-y: hidden;
  position: relative;
}
.tab-nav-scroll::-webkit-scrollbar { height: 0; }

.tab-nav-list {
  display: flex;
  align-items: stretch;
  height: 100%;
  padding: 0 4px;
}

.tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 14px;
  font-size: 13px;
  color: var(--text-secondary, #666);
  cursor: pointer;
  white-space: nowrap;
  position: relative;
  transition: all 0.15s;
  user-select: none;
  border-radius: 6px 6px 0 0;
  margin: 4px 2px 0;
}
.tab-item:hover {
  color: var(--text-primary, #333);
  background: var(--hover-bg, rgba(0,0,0,0.04));
}
.tab-item.active {
  color: var(--primary, #4f46e5);
  background: var(--primary-suppl, rgba(79,70,229,0.06));
}
.tab-item.pinned {
  padding-right: 8px;
}
.tab-item.drag-over {
  border-left: 2px solid var(--primary, #4f46e5);
}

.tab-icon {
  font-size: 15px;
}

.tab-title {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tab-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 4px;
  color: var(--text-muted, #999);
  transition: all 0.15s;
}
.tab-close:hover {
  background: rgba(0,0,0,0.1);
  color: var(--text-primary, #333);
}

.tab-bar {
  position: absolute;
  bottom: 0;
  height: 2px;
  background: var(--primary, #4f46e5);
  border-radius: 1px 1px 0 0;
  transition: left 0.25s cubic-bezier(0.4, 0, 0.2, 1), width 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.tab-suffix {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 0 8px;
  flex-shrink: 0;
}

.suffix-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--text-secondary, #666);
  cursor: pointer;
  transition: all 0.15s;
}
.suffix-btn:hover {
  background: var(--hover-bg, rgba(0,0,0,0.06));
  color: var(--text-primary, #333);
}

.context-menu {
  position: fixed;
  z-index: 1000;
  background: var(--card-bg, #fff);
  border: 1px solid var(--border-color, #e8e8e8);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  padding: 4px;
  min-width: 140px;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  font-size: 13px;
  color: var(--text-primary, #333);
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.15s;
}
.menu-item:hover {
  background: var(--hover-bg, rgba(0,0,0,0.06));
}

.menu-divider {
  height: 1px;
  background: var(--border-color, #e8e8e8);
  margin: 4px 8px;
}
</style>