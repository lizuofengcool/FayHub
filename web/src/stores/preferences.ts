import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export type LayoutMode = 'side' | 'top' | 'mix' | 'mix-double' | 'mix-sidebar' | 'full'
export type ThemeMode = 'light' | 'dark' | 'auto'
export type ThemeType = 'pure' | 'skin'
export type TabStyle = 'google' | 'card' | 'smart'
export type CollapseMode = 'expand' | 'collapse' | 'side'
export type AnimationType = 'fade' | 'slide' | 'zoom'
export type PrefPosition = 'fixed-bottom-right' | 'fixed-bottom-left' | 'header'

export interface Preferences {
  themeColor: string
  themeMode: ThemeMode
  themeType: ThemeType
  skin: string
  layoutMode: LayoutMode
  defaultLayout: LayoutMode
  layoutTheme: 'light' | 'dark'
  sidebarTheme: 'light' | 'dark'
  headerTheme: 'light' | 'dark'
  radius: string
  grayMode: boolean
  colorBlind: boolean

  headerVisible: boolean
  fixedHeader: boolean
  headerHeight: number
  breadcrumbVisible: boolean
  breadcrumbIcon: boolean
  tabbarVisible: boolean
  tabbarIcon: boolean
  tabbarHeight: number
  tabbarStyle: TabStyle
  sidebarVisible: boolean
  accordion: boolean
  collapseMode: CollapseMode
  showMenuName: boolean
  sidebarCollapsed: number
  sidebarWidth: number
  collapseWidth: number
  footerVisible: boolean
  fixedFooter: boolean
  footerHeight: number
  gap: boolean
  menugroup: boolean
  menudivider: boolean

  i18n: boolean
  fullscreenBtn: boolean
  refresh: boolean
  themeSwitchBtn: boolean
  sidebarCollapseBtn: boolean
  notification: boolean
  prefPosition: PrefPosition
  progressBar: boolean
  animationType: AnimationType

  logo: boolean
  largeLogo: boolean
  hideBorder: boolean
  dynamicTitle: boolean
  watermark: boolean
  headerTransparent: boolean
  sidebarTransparent: boolean
}

const STORAGE_KEY = 'fayhub_preferences'

const defaultPreferences: Preferences = {
  themeColor: '#2d8cf0',
  themeMode: 'light',
  themeType: 'pure',
  skin: 'blue-sky',
  layoutMode: 'side',
  defaultLayout: 'side',
  layoutTheme: 'light',
  sidebarTheme: 'dark',
  headerTheme: 'light',
  radius: '0.5',
  grayMode: false,
  colorBlind: false,

  headerVisible: true,
  fixedHeader: true,
  headerHeight: 56,
  breadcrumbVisible: true,
  breadcrumbIcon: true,
  tabbarVisible: true,
  tabbarIcon: true,
  tabbarHeight: 47,
  tabbarStyle: 'google',
  sidebarVisible: true,
  accordion: false,
  collapseMode: 'expand',
  showMenuName: false,
  sidebarCollapsed: 90,
  sidebarWidth: 224,
  collapseWidth: 64,
  footerVisible: true,
  fixedFooter: true,
  footerHeight: 70,
  gap: true,
  menugroup: true,
  menudivider: true,

  i18n: true,
  fullscreenBtn: true,
  refresh: true,
  themeSwitchBtn: true,
  sidebarCollapseBtn: true,
  notification: true,
  prefPosition: 'fixed-bottom-right',
  progressBar: true,
  animationType: 'fade',

  logo: true,
  largeLogo: false,
  hideBorder: false,
  dynamicTitle: false,
  watermark: false,
  headerTransparent: false,
  sidebarTransparent: false,
}

function loadFromStorage(): Preferences {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved) {
      return { ...defaultPreferences, ...JSON.parse(saved) }
    }
  } catch { /* ignore */ }
  return { ...defaultPreferences }
}

function lightenColor(hex: string, amount: number): string {
  const num = parseInt(hex.replace('#', ''), 16)
  const r = Math.min(255, (num >> 16) + 255 * amount)
  const g = Math.min(255, ((num >> 8) & 0x00FF) + 255 * amount)
  const b = Math.min(255, (num & 0x0000FF) + 255 * amount)
  return '#' + (0x1000000 + (Math.round(r) << 16) + (Math.round(g) << 8) + Math.round(b)).toString(16).slice(1)
}

function darkenColor(hex: string, amount: number): string {
  const num = parseInt(hex.replace('#', ''), 16)
  const r = Math.max(0, (num >> 16) - 255 * amount)
  const g = Math.max(0, ((num >> 8) & 0x00FF) - 255 * amount)
  const b = Math.max(0, (num & 0x0000FF) - 255 * amount)
  return '#' + (0x1000000 + (Math.round(r) << 16) + (Math.round(g) << 8) + Math.round(b)).toString(16).slice(1)
}

const skinColors: Record<string, string> = {
  'blue-sky': '#2d8cf0',
  'blue-xmas': '#2d6cb4',
  'colorful-mica': '#667eea',
  'pink-romance': '#f5576c',
  'emerald-green': '#11998e',
  'flowing-light': '#f7971e',
  'orange-bubble': '#ff9a9e',
  'deepspace-jelly': '#1a1a4e',
  'starlight-neon': '#fc00ff',
}

export const usePreferencesStore = defineStore('preferences', () => {
  const prefs = ref<Preferences>(loadFromStorage())

  function save() {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(prefs.value))
  }

  function applyAll() {
    const html = document.documentElement
    const p = prefs.value

    html.style.setProperty('--primary', p.themeColor)
    html.style.setProperty('--primary-hover', lightenColor(p.themeColor, 0.1))
    html.style.setProperty('--primary-active', darkenColor(p.themeColor, 0.05))
    html.style.setProperty('--primary-suppl', p.themeColor + '1a')
    html.style.setProperty('--sidebar-active-bg', p.themeColor + '26')
    html.style.setProperty('--tab-active-color', p.themeColor)
    html.style.setProperty('--body-bg-tint', p.themeColor + '08')

    html.removeAttribute('data-theme')
    if (p.themeMode === 'dark') {
      html.setAttribute('data-theme', 'dark')
    } else if (p.themeMode === 'auto') {
      if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
        html.setAttribute('data-theme', 'dark')
      }
    }

    html.setAttribute('data-theme-type', p.themeType)

    html.removeAttribute('data-layout')
    if (p.layoutMode !== 'side') {
      html.setAttribute('data-layout', p.layoutMode)
    }

    if (p.sidebarTheme === 'light') {
      html.setAttribute('data-nav-style', 'light-sidebar')
    } else {
      html.removeAttribute('data-nav-style')
    }

    if (p.headerTheme === 'dark') {
      html.setAttribute('data-header-style', 'dark')
    } else {
      html.removeAttribute('data-header-style')
    }

    const remVal = parseFloat(p.radius)
    html.style.setProperty('--radius', (remVal * 16) + 'px')
    html.style.setProperty('--radius-sm', (remVal * 12) + 'px')

    document.body.classList.toggle('grayscale-mode', p.grayMode)
    document.body.classList.toggle('color-blind-mode', p.colorBlind)

    html.style.setProperty('--nav-height', p.headerHeight + 'px')
    html.style.setProperty('--tabbar-height', p.tabbarHeight + 'px')
    html.style.setProperty('--sidebar-collapsed', p.sidebarCollapsed + 'px')
    html.style.setProperty('--sidebar-width', p.sidebarWidth + 'px')
    html.style.setProperty('--collapse-width', p.collapseWidth + 'px')
    html.style.setProperty('--footer-height', p.footerHeight + 'px')

    document.body.classList.toggle('header-hidden', !p.headerVisible)
    document.body.classList.toggle('header-fixed', p.fixedHeader)
    document.body.classList.toggle('footer-hidden', !p.footerVisible)
    document.body.classList.toggle('footer-fixed', p.fixedFooter)
    document.body.classList.toggle('breadcrumb-hidden', !p.breadcrumbVisible)
    document.body.classList.toggle('show-menu-name', p.showMenuName)
    document.body.classList.toggle('accordion-mode', p.accordion)
    document.body.classList.toggle('no-menu-group', !p.menugroup)
    document.body.classList.toggle('header-transparent', p.headerTransparent)
    document.body.classList.toggle('sidebar-transparent', p.sidebarTransparent)
    document.body.classList.toggle('hide-layout-border', p.hideBorder)
    document.body.classList.toggle('no-progress-bar', !p.progressBar)

    document.body.setAttribute('data-tab-style', p.tabbarStyle)
    document.body.setAttribute('data-animation', p.animationType)
  }

  function updateThemeColor(color: string) {
    prefs.value.themeColor = color
    applyAll()
    save()
  }

  function setThemeMode(mode: ThemeMode) {
    prefs.value.themeMode = mode
    applyAll()
    save()
  }

  function setLayoutMode(mode: LayoutMode) {
    if (mode === 'full' && prefs.value.layoutMode !== 'full') {
      prefs.value.defaultLayout = prefs.value.layoutMode
    }
    prefs.value.layoutMode = mode
    applyAll()
    save()
  }

  function setThemeType(type: ThemeType) {
    prefs.value.themeType = type
    if (type === 'pure') {
      prefs.value.headerTransparent = false
      prefs.value.sidebarTransparent = false
    } else {
      prefs.value.headerTransparent = true
      prefs.value.sidebarTransparent = true
    }
    applyAll()
    save()
  }

  function setSkin(skin: string) {
    prefs.value.skin = skin
    const color = skinColors[skin] || '#2d8cf0'
    prefs.value.themeColor = color
    applyAll()
    save()
  }

  function setRadius(val: string) {
    prefs.value.radius = val
    applyAll()
    save()
  }

  function toggleGrayMode() {
    prefs.value.grayMode = !prefs.value.grayMode
    applyAll()
    save()
  }

  function toggleColorBlind() {
    prefs.value.colorBlind = !prefs.value.colorBlind
    applyAll()
    save()
  }

  function toggleSwitch(key: keyof Preferences) {
    const val = prefs.value[key]
    if (typeof val === 'boolean') {
      ;(prefs.value as any)[key] = !val
      applyAll()
      save()
    }
  }

  function setNumber(key: keyof Preferences, val: number) {
    ;(prefs.value as any)[key] = val
    applyAll()
    save()
  }

  function setSelect(key: keyof Preferences, val: string) {
    ;(prefs.value as any)[key] = val
    applyAll()
    save()
  }

  function setSidebarTheme(theme: 'light' | 'dark') {
    prefs.value.sidebarTheme = theme
    applyAll()
    save()
  }

  function setHeaderTheme(theme: 'light' | 'dark') {
    prefs.value.headerTheme = theme
    applyAll()
    save()
  }

  function resetAll() {
    prefs.value = { ...defaultPreferences }
    applyAll()
    save()
  }

  function getConfigJSON(): string {
    return JSON.stringify(prefs.value, null, 2)
  }

  watch(prefs, save, { deep: true })

  return {
    prefs,
    applyAll,
    updateThemeColor,
    setThemeMode,
    setLayoutMode,
    setThemeType,
    setSkin,
    setRadius,
    toggleGrayMode,
    toggleColorBlind,
    toggleSwitch,
    setNumber,
    setSelect,
    setSidebarTheme,
    setHeaderTheme,
    resetAll,
    getConfigJSON,
    skinColors,
  }
})
