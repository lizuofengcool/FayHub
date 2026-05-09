import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  const isDark = ref(false)
  const sidebarCollapsed = ref(false)
  const showTabs = ref(true)
  const fixedHeader = ref(true)

  // 从 localStorage 读取主题设置
  const initTheme = () => {
    const saved = localStorage.getItem('fayhub_theme')
    if (saved) {
      try {
        const config = JSON.parse(saved)
        isDark.value = config.isDark || false
        sidebarCollapsed.value = config.sidebarCollapsed || false
        showTabs.value = config.showTabs !== false
        fixedHeader.value = config.fixedHeader !== false
      } catch {
        // ignore
      }
    }
    applyTheme()
  }

  const applyTheme = () => {
    if (isDark.value) {
      document.documentElement.setAttribute('data-theme', 'dark')
    } else {
      document.documentElement.removeAttribute('data-theme')
    }
  }

  const saveTheme = () => {
    localStorage.setItem('fayhub_theme', JSON.stringify({
      isDark: isDark.value,
      sidebarCollapsed: sidebarCollapsed.value,
      showTabs: showTabs.value,
      fixedHeader: fixedHeader.value
    }))
  }

  function toggleTheme() {
    isDark.value = !isDark.value
    applyTheme()
    saveTheme()
  }

  function setTheme(dark: boolean) {
    isDark.value = dark
    applyTheme()
    saveTheme()
  }

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
    saveTheme()
  }

  // 监听变化自动保存
  watch([isDark, sidebarCollapsed, showTabs, fixedHeader], saveTheme, { deep: true })

  return {
    isDark,
    sidebarCollapsed,
    showTabs,
    fixedHeader,
    toggleTheme,
    setTheme,
    toggleSidebar,
    initTheme
  }
})
