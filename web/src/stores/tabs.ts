import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { RouteLocationNormalized } from 'vue-router'

export interface TabItem {
  path: string
  title: string
  name?: string
  closable?: boolean
}

export const useTabsStore = defineStore('tabs', () => {
  const tabs = ref<TabItem[]>([
    { path: '/dashboard', title: '首页', closable: false }
  ])
  const activeTab = ref<string>('/dashboard')

  function addTab(route: RouteLocationNormalized) {
    const exists = tabs.value.find(tab => tab.path === route.path)
    if (!exists && route.meta?.title) {
      tabs.value.push({
        path: route.path,
        title: route.meta.title as string,
        name: route.name as string,
        closable: route.path !== '/dashboard'
      })
    }
    if (route.path !== activeTab.value) {
      activeTab.value = route.path
    }
  }

  function closeTab(path: string) {
    const index = tabs.value.findIndex(tab => tab.path === path)
    if (index > -1 && tabs.value[index].closable !== false) {
      tabs.value.splice(index, 1)
      if (activeTab.value === path) {
        const nextTab = tabs.value[index] || tabs.value[index - 1]
        if (nextTab) {
          activeTab.value = nextTab.path
        }
      }
    }
    return activeTab.value
  }

  function closeOtherTabs() {
    tabs.value = tabs.value.filter(tab => !tab.closable)
  }

  function closeAllTabs() {
    tabs.value = tabs.value.filter(tab => !tab.closable)
    activeTab.value = '/dashboard'
  }

  return {
    tabs,
    activeTab,
    addTab,
    closeTab,
    closeOtherTabs,
    closeAllTabs
  }
})