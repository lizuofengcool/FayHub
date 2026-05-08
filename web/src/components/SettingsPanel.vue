<template>
  <Teleport to="body">
    <div class="settings-overlay" :class="{ show: visible }" @click="$emit('close')"></div>
    <div class="settings-panel" :class="{ show: visible }">
      <div class="settings-header">
        <h3>偏好配置</h3>
        <button class="settings-close" @click="$emit('close')">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 6 6 18M6 6l12 12"/></svg>
        </button>
      </div>

      <div class="settings-tabs">
        <button
          v-for="(tab, idx) in tabs"
          :key="idx"
          class="settings-tab"
          :class="{ active: activeTab === idx }"
          @click="activeTab = idx"
        >{{ tab }}</button>
      </div>

      <div class="settings-body">
        <div v-show="activeTab === 0" class="settings-tab-content">
          <div class="settings-section">
            <h4>主题模式</h4>
            <div class="segment-tabs">
              <button
                v-for="opt in themeModeOptions"
                :key="opt.value"
                class="seg-tab"
                :class="{ active: prefs.themeMode === opt.value }"
                @click="store.setThemeMode(opt.value)"
              >{{ opt.label }}</button>
            </div>
          </div>

          <div class="settings-section">
            <h4>主题色</h4>
            <div class="color-presets-grid">
              <button
                v-for="c in allPresetColors"
                :key="c"
                class="color-dot"
                :class="{ active: prefs.themeColor === c }"
                :style="{ background: c }"
                @click="store.updateThemeColor(c)"
              ></button>
              <div class="color-custom">
                <input type="color" :value="prefs.themeColor" @change="store.updateThemeColor(($event.target as HTMLInputElement).value)">
              </div>
            </div>
          </div>

          <div class="settings-section">
            <h4>侧边栏主题</h4>
            <div class="radio-group">
              <label
                class="radio-btn"
                :class="{ active: prefs.sidebarTheme === 'dark' }"
                @click="store.setSidebarTheme('dark')"
              >深色</label>
              <label
                class="radio-btn"
                :class="{ active: prefs.sidebarTheme === 'light' }"
                @click="store.setSidebarTheme('light')"
              >浅色</label>
            </div>
          </div>

          <div class="settings-section">
            <h4>顶栏主题</h4>
            <div class="radio-group">
              <label
                class="radio-btn"
                :class="{ active: prefs.headerTheme === 'light' }"
                @click="store.setHeaderTheme('light')"
              >浅色</label>
              <label
                class="radio-btn"
                :class="{ active: prefs.headerTheme === 'dark' }"
                @click="store.setHeaderTheme('dark')"
              >深色</label>
            </div>
          </div>

          <div class="settings-section">
            <h4>圆角</h4>
            <div class="radio-group">
              <label
                v-for="r in radiusOptions"
                :key="r"
                class="radio-btn"
                :class="{ active: prefs.radius === r }"
                @click="store.setRadius(r)"
              >{{ r }}</label>
            </div>
          </div>

          <div class="settings-section">
            <div class="settings-row">
              <label>色弱模式</label>
              <button class="toggle" :class="{ on: prefs.colorBlind }" @click="store.toggleColorBlind()"></button>
            </div>
          </div>

          <div class="settings-section">
            <div class="settings-row">
              <label>灰色模式</label>
              <button class="toggle" :class="{ on: prefs.grayMode }" @click="store.toggleGrayMode()"></button>
            </div>
          </div>
        </div>

        <div v-show="activeTab === 1" class="settings-tab-content">
          <div class="settings-section">
            <h4>布局模式</h4>
            <div class="layout-cards">
              <div
                v-for="layout in layoutOptions"
                :key="layout.value"
                class="layout-card"
                :class="{ active: prefs.layoutMode === layout.value }"
                @click="store.setLayoutMode(layout.value)"
              >
                <div class="layout-preview">
                  <svg viewBox="0 0 60 40" width="60" height="40">
                    <!-- 垂直布局 side: 左侧边栏 + 右侧顶栏+内容 -->
                    <template v-if="layout.value === 'side'">
                      <rect x="0" y="0" width="12" height="40" rx="2" :fill="primaryColor" />
                      <rect x="14" y="2" width="44" height="8" rx="1.5" fill="#e0e0e0" />
                      <rect x="14" y="13" width="44" height="25" rx="1.5" fill="#f0f0f0" />
                    </template>
                    <!-- 水平布局 top: 顶部导航 + 下方内容 -->
                    <template v-else-if="layout.value === 'top'">
                      <rect x="0" y="0" width="60" height="8" rx="2" :fill="primaryColor" />
                      <rect x="2" y="12" width="56" height="26" rx="1.5" fill="#f0f0f0" />
                    </template>
                    <!-- 双栏布局 mix: 双左侧边栏 + 右侧顶栏+内容 -->
                    <template v-else-if="layout.value === 'mix'">
                      <rect x="0" y="0" width="8" height="40" rx="2" :fill="primaryColor" />
                      <rect x="10" y="0" width="12" height="40" rx="2" :fill="primaryColor" opacity="0.35" />
                      <rect x="24" y="2" width="34" height="10" rx="1.5" fill="#e0e0e0" />
                      <rect x="24" y="15" width="34" height="23" rx="1.5" fill="#f0f0f0" />
                    </template>
                    <!-- 混合双栏 mix-double: 双左侧边栏 + 右侧顶栏(主色)+内容 -->
                    <template v-else-if="layout.value === 'mix-double'">
                      <rect x="0" y="0" width="8" height="40" rx="2" :fill="primaryColor" />
                      <rect x="10" y="0" width="12" height="40" rx="2" :fill="primaryColor" opacity="0.35" />
                      <rect x="24" y="2" width="34" height="10" rx="1.5" :fill="primaryColor" />
                      <rect x="24" y="15" width="34" height="23" rx="1.5" fill="#f0f0f0" />
                    </template>
                    <!-- 内容全屏 full: 双顶栏 + 内容 -->
                    <template v-else-if="layout.value === 'full'">
                      <rect x="0" y="0" width="28" height="10" rx="1.5" fill="#e0e0e0" />
                      <rect x="32" y="0" width="28" height="10" rx="1.5" fill="#e0e0e0" />
                      <rect x="0" y="14" width="60" height="24" rx="1.5" fill="#f0f0f0" />
                    </template>
                    <!-- 混合侧边栏 mix-sidebar: 左侧边栏 + 右侧顶栏(主色)+内容 -->
                    <template v-else-if="layout.value === 'mix-sidebar'">
                      <rect x="0" y="0" width="12" height="40" rx="2" :fill="primaryColor" />
                      <rect x="14" y="2" width="44" height="10" rx="1.5" :fill="primaryColor" />
                      <rect x="14" y="15" width="44" height="23" rx="1.5" fill="#f0f0f0" />
                    </template>
                  </svg>
                </div>
                <span>{{ layout.label }}</span>
              </div>
            </div>
          </div>

          <div class="settings-section">
            <div class="settings-row">
              <label>间隙布局</label>
              <button class="toggle" :class="{ on: prefs.gap }" @click="store.toggleSwitch('gap')"></button>
            </div>
          </div>

          <div class="settings-section">
            <div class="settings-row">
              <label>菜单分组</label>
              <button class="toggle" :class="{ on: prefs.menugroup }" @click="store.toggleSwitch('menugroup')"></button>
            </div>
          </div>

          <div class="settings-section">
            <div class="settings-row">
              <label>菜单分割线</label>
              <button class="toggle" :class="{ on: prefs.menudivider }" @click="store.toggleSwitch('menudivider')"></button>
            </div>
          </div>

          <div class="settings-section section-block">
            <h4>顶栏</h4>
            <div class="settings-row">
              <label>显示顶栏</label>
              <button class="toggle" :class="{ on: prefs.headerVisible }" @click="store.toggleSwitch('headerVisible')"></button>
            </div>
            <div class="settings-row">
              <label>固定顶栏</label>
              <button class="toggle" :class="{ on: prefs.fixedHeader }" @click="store.toggleSwitch('fixedHeader')"></button>
            </div>
            <div class="settings-row">
              <label>顶栏高度</label>
              <div class="num-group">
                <button class="num-btn" @click="adjustNumber('headerHeight', -1)">−</button>
                <input type="number" class="num-input" :value="prefs.headerHeight" min="40" max="80" @change="store.setNumber('headerHeight', Number(($event.target as HTMLInputElement).value))">
                <button class="num-btn" @click="adjustNumber('headerHeight', 1)">+</button>
              </div>
            </div>
          </div>

          <div class="settings-section section-block">
            <h4>面包屑</h4>
            <div class="settings-row">
              <label>显示面包屑</label>
              <button class="toggle" :class="{ on: prefs.breadcrumbVisible }" @click="store.toggleSwitch('breadcrumbVisible')"></button>
            </div>
            <div class="settings-row">
              <label>显示面包屑图标</label>
              <button class="toggle" :class="{ on: prefs.breadcrumbIcon }" @click="store.toggleSwitch('breadcrumbIcon')"></button>
            </div>
          </div>

          <div class="settings-section section-block">
            <h4>页签</h4>
            <div class="settings-row">
              <label>显示页签</label>
              <button class="toggle" :class="{ on: prefs.tabbarVisible }" @click="store.toggleSwitch('tabbarVisible')"></button>
            </div>
            <div class="settings-row">
              <label>显示页签图标</label>
              <button class="toggle" :class="{ on: prefs.tabbarIcon }" @click="store.toggleSwitch('tabbarIcon')"></button>
            </div>
            <div class="settings-row">
              <label>页签高度</label>
              <div class="num-group">
                <button class="num-btn" @click="adjustNumber('tabbarHeight', -1)">−</button>
                <input type="number" class="num-input" :value="prefs.tabbarHeight" min="30" max="60" @change="store.setNumber('tabbarHeight', Number(($event.target as HTMLInputElement).value))">
                <button class="num-btn" @click="adjustNumber('tabbarHeight', 1)">+</button>
              </div>
            </div>
            <div class="settings-row">
              <label>页签风格</label>
              <select :value="prefs.tabbarStyle" @change="store.setSelect('tabbarStyle', ($event.target as HTMLSelectElement).value)">
                <option value="google">谷歌</option>
                <option value="card">卡片</option>
                <option value="smart">灵动</option>
              </select>
            </div>
          </div>

          <div class="settings-section section-block">
            <h4>侧边栏</h4>
            <div class="settings-row">
              <label>显示侧边栏</label>
              <button class="toggle" :class="{ on: prefs.sidebarVisible }" @click="store.toggleSwitch('sidebarVisible')"></button>
            </div>
            <div class="settings-row">
              <label>手风琴模式</label>
              <button class="toggle" :class="{ on: prefs.accordion }" @click="store.toggleSwitch('accordion')"></button>
            </div>
            <div class="settings-row">
              <label>折叠模式</label>
              <select :value="prefs.collapseMode" @change="store.setSelect('collapseMode', ($event.target as HTMLSelectElement).value)">
                <option value="expand">展开</option>
                <option value="collapse">折叠</option>
                <option value="side">侧边栏</option>
              </select>
            </div>
            <div class="settings-row">
              <label>折叠后显示菜单名</label>
              <button class="toggle" :class="{ on: prefs.showMenuName }" @click="store.toggleSwitch('showMenuName')"></button>
            </div>
            <div class="settings-row">
              <label>侧边栏宽度</label>
              <div class="num-group">
                <button class="num-btn" @click="adjustNumber('sidebarWidth', -1)">−</button>
                <input type="number" class="num-input" :value="prefs.sidebarWidth" min="180" max="320" @change="store.setNumber('sidebarWidth', Number(($event.target as HTMLInputElement).value))">
                <button class="num-btn" @click="adjustNumber('sidebarWidth', 1)">+</button>
              </div>
            </div>
            <div class="settings-row">
              <label>折叠宽度</label>
              <div class="num-group">
                <button class="num-btn" @click="adjustNumber('collapseWidth', -1)">−</button>
                <input type="number" class="num-input" :value="prefs.collapseWidth" min="40" max="100" @change="store.setNumber('collapseWidth', Number(($event.target as HTMLInputElement).value))">
                <button class="num-btn" @click="adjustNumber('collapseWidth', 1)">+</button>
              </div>
            </div>
          </div>

          <div class="settings-section section-block">
            <h4>底部</h4>
            <div class="settings-row">
              <label>显示底部</label>
              <button class="toggle" :class="{ on: prefs.footerVisible }" @click="store.toggleSwitch('footerVisible')"></button>
            </div>
            <div class="settings-row">
              <label>固定底部</label>
              <button class="toggle" :class="{ on: prefs.fixedFooter }" @click="store.toggleSwitch('fixedFooter')"></button>
            </div>
            <div class="settings-row">
              <label>底部高度</label>
              <div class="num-group">
                <button class="num-btn" @click="adjustNumber('footerHeight', -1)">−</button>
                <input type="number" class="num-input" :value="prefs.footerHeight" min="40" max="120" @change="store.setNumber('footerHeight', Number(($event.target as HTMLInputElement).value))">
                <button class="num-btn" @click="adjustNumber('footerHeight', 1)">+</button>
              </div>
            </div>
          </div>

          <div class="settings-section section-block">
            <h4>小部件</h4>
            <div class="settings-row">
              <label>多语言</label>
              <button class="toggle" :class="{ on: prefs.i18n }" @click="store.toggleSwitch('i18n')"></button>
            </div>
            <div class="settings-row">
              <label>全屏</label>
              <button class="toggle" :class="{ on: prefs.fullscreenBtn }" @click="store.toggleSwitch('fullscreenBtn')"></button>
            </div>
            <div class="settings-row">
              <label>刷新</label>
              <button class="toggle" :class="{ on: prefs.refresh }" @click="store.toggleSwitch('refresh')"></button>
            </div>
            <div class="settings-row">
              <label>主题切换</label>
              <button class="toggle" :class="{ on: prefs.themeSwitchBtn }" @click="store.toggleSwitch('themeSwitchBtn')"></button>
            </div>
            <div class="settings-row">
              <label>侧边栏折叠</label>
              <button class="toggle" :class="{ on: prefs.sidebarCollapseBtn }" @click="store.toggleSwitch('sidebarCollapseBtn')"></button>
            </div>
            <div class="settings-row">
              <label>消息通知</label>
              <button class="toggle" :class="{ on: prefs.notification }" @click="store.toggleSwitch('notification')"></button>
            </div>
          </div>

          <div class="settings-section section-block">
            <h4>动画</h4>
            <div class="settings-row">
              <label>页面切换进度条</label>
              <button class="toggle" :class="{ on: prefs.progressBar }" @click="store.toggleSwitch('progressBar')"></button>
            </div>
            <div class="settings-row">
              <label>动画类型</label>
              <select :value="prefs.animationType" @change="store.setSelect('animationType', ($event.target as HTMLSelectElement).value)">
                <option value="fade">消退</option>
                <option value="slide">滑动</option>
                <option value="zoom">缩放</option>
              </select>
            </div>
          </div>

          <div class="settings-section section-block">
            <h4>其他</h4>
            <div class="settings-row">
              <label>显示 Logo</label>
              <button class="toggle" :class="{ on: prefs.logo }" @click="store.toggleSwitch('logo')"></button>
            </div>
            <div class="settings-row">
              <label>放大 Logo</label>
              <button class="toggle" :class="{ on: prefs.largeLogo }" @click="store.toggleSwitch('largeLogo')"></button>
            </div>
            <div class="settings-row">
              <label>隐藏布局边框</label>
              <button class="toggle" :class="{ on: prefs.hideBorder }" @click="store.toggleSwitch('hideBorder')"></button>
            </div>
            <div class="settings-row">
              <label>动态标题</label>
              <button class="toggle" :class="{ on: prefs.dynamicTitle }" @click="store.toggleSwitch('dynamicTitle')"></button>
            </div>
            <div class="settings-row">
              <label>开启水印</label>
              <button class="toggle" :class="{ on: prefs.watermark }" @click="store.toggleSwitch('watermark')"></button>
            </div>
          </div>
        </div>

        <div v-show="activeTab === 2" class="settings-tab-content">
          <div class="settings-section" style="text-align:center;">
            <div class="radio-group" style="justify-content:center;margin-bottom:16px;">
              <label
                class="radio-btn"
                :class="{ active: prefs.themeType === 'pure' }"
                @click="store.setThemeType('pure')"
              >纯色主题</label>
              <label
                class="radio-btn"
                :class="{ active: prefs.themeType === 'skin' }"
                @click="store.setThemeType('skin')"
              >主题皮肤</label>
            </div>

            <div v-if="prefs.themeType === 'pure'">
              <div class="pure-color-section">
                <ColorPickerPopup v-model="prefs.themeColor" @update:model-value="store.updateThemeColor($event)" />
              </div>
            </div>

            <div v-else class="skin-grid">
              <div
                v-for="skin in skinList"
                :key="skin.key"
                class="skin-card"
                :class="{ active: prefs.skin === skin.key }"
                @click="store.setSkin(skin.key)"
              >
                <div class="skin-preview" :style="{ background: skin.gradient }"></div>
                <span>{{ skin.label }}</span>
              </div>
            </div>
          </div>

          <div class="settings-section">
            <div class="settings-row">
              <label>顶部透明</label>
              <button
                class="toggle"
                :class="{ on: prefs.headerTransparent }"
                :disabled="prefs.themeType === 'pure'"
                @click="store.toggleSwitch('headerTransparent')"
              ></button>
            </div>
          </div>
          <div class="settings-section">
            <div class="settings-row">
              <label>侧边栏透明</label>
              <button
                class="toggle"
                :class="{ on: prefs.sidebarTransparent }"
                :disabled="prefs.themeType === 'pure'"
                @click="store.toggleSwitch('sidebarTransparent')"
              ></button>
            </div>
          </div>
        </div>
      </div>

      <div class="settings-footer">
        <button class="btn btn-sm" @click="copyConfig">复制配置</button>
        <button class="btn btn-sm btn-primary" @click="resetConfig">重置配置</button>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { usePreferencesStore } from '@/stores/preferences'
import type { LayoutMode } from '@/stores/preferences'
import ColorPickerPopup from './ColorPickerPopup.vue'

defineProps<{ visible: boolean }>()
defineEmits<{ close: [] }>()

const store = usePreferencesStore()
const prefs = computed(() => store.prefs)
const primaryColor = computed(() => prefs.value.themeColor || '#2d8cf0')

const activeTab = ref(0)
const tabs = ['主题设置', '框架布局', '主题风格']

const themeModeOptions = [
  { value: 'light' as const, label: '☀ 浅色' },
  { value: 'dark' as const, label: '🌙 深色' },
  { value: 'auto' as const, label: '💻 自动' },
]

const allPresetColors = [
  '#2d8cf0', '#18a058', '#f0a020', '#d03050', '#2080f0',
  '#667eea', '#f5576c', '#11998e', '#f7971e', '#1a1a4e',
  '#ff6b6b', '#4ecdc4', '#45b7d1', '#96ceb4', '#ffeaa7',
  '#dda0dd', '#98d8c8', '#f7dc6f', '#bb8fce',
]

const radiusOptions = ['0', '0.25', '0.5', '0.75', '1']

const layoutOptions: { value: LayoutMode; label: string }[] = [
  { value: 'side', label: '垂直布局' },
  { value: 'top', label: '水平布局' },
  { value: 'mix', label: '双栏布局' },
  { value: 'mix-double', label: '混合双栏' },
  { value: 'mix-sidebar', label: '混合侧边栏' },
  { value: 'full', label: '内容全屏' },
]

const skinList = [
  { key: 'blue-sky', label: '蓝色天空', gradient: 'linear-gradient(135deg,#2d8cf0,#58adfc)' },
  { key: 'blue-xmas', label: '蓝色圣诞', gradient: 'linear-gradient(135deg,#1a3a5c,#2d6cb4)' },
  { key: 'colorful-mica', label: '彩色云母', gradient: 'linear-gradient(135deg,#667eea,#764ba2)' },
  { key: 'pink-romance', label: '粉色浪漫', gradient: 'linear-gradient(135deg,#f093fb,#f5576c)' },
  { key: 'emerald-green', label: '翡翠绿峰', gradient: 'linear-gradient(135deg,#11998e,#38ef7d)' },
  { key: 'flowing-light', label: '流光溢彩', gradient: 'linear-gradient(135deg,#f7971e,#ffd200)' },
  { key: 'orange-bubble', label: '香橙泡泡', gradient: 'linear-gradient(135deg,#ff9a9e,#fecfef)' },
  { key: 'deepspace-jelly', label: '深空水母', gradient: 'linear-gradient(135deg,#0c0c1d,#1a1a4e)' },
  { key: 'starlight-neon', label: '星光霓虹', gradient: 'linear-gradient(135deg,#fc00ff,#00dbde)' },
]

function adjustNumber(key: string, delta: number) {
  const current = (prefs.value as any)[key] as number
  store.setNumber(key as any, current + delta)
}

function copyConfig() {
  const text = store.getConfigJSON()
  if (navigator.clipboard) {
    navigator.clipboard.writeText(text).then(() => alert('配置已复制到剪贴板'))
  } else {
    const ta = document.createElement('textarea')
    ta.value = text
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
    alert('配置已复制到剪贴板')
  }
}

function resetConfig() {
  if (confirm('确定要重置所有偏好配置吗？')) {
    store.resetAll()
  }
}
</script>

<style scoped>
.settings-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.3);
  z-index: 1000;
  opacity: 0;
  transition: opacity 0.3s;
  pointer-events: none;
}
.settings-overlay.show {
  opacity: 1;
  pointer-events: auto;
}

.settings-panel {
  position: fixed;
  top: 0;
  right: 0;
  width: 340px;
  height: 100vh;
  background: #fff;
  z-index: 1001;
  box-shadow: -4px 0 24px rgba(0, 0, 0, 0.08);
  transform: translateX(100%);
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  flex-direction: column;
}
.settings-panel.show {
  transform: translateX(0);
}

.settings-header {
  padding: 16px 20px;
  border-bottom: 0.8px solid #eee;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}
.settings-header h3 {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  color: #333;
}

.settings-close {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border-radius: 6px;
  border: none;
  background: none;
  color: #999;
  transition: all 0.15s;
}
.settings-close:hover {
  background: rgba(0, 0, 0, 0.04);
  color: #333;
}

.settings-tabs {
  display: flex;
  border-bottom: 0.8px solid #eee;
  padding: 0 20px;
  gap: 0;
  flex-shrink: 0;
}
.settings-tab {
  padding: 10px 16px;
  cursor: pointer;
  font-size: 14px;
  color: #999;
  border-bottom: 2px solid transparent;
  transition: all 0.15s;
  background: none;
  border-top: none;
  border-left: none;
  border-right: none;
}
.settings-tab:hover { color: #666; }
.settings-tab.active {
  color: var(--primary, #2d8cf0);
  border-bottom-color: var(--primary, #2d8cf0);
}

.settings-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
}

.settings-section {
  margin-bottom: 20px;
}
.settings-section h4 {
  font-size: 13px;
  color: #999;
  margin-bottom: 10px;
  font-weight: 500;
  text-transform: none;
}

.settings-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 0;
}
.settings-row label {
  font-size: 14px;
  color: #333;
}
.settings-row select {
  padding: 4px 8px;
  border: 0.8px solid #ddd;
  border-radius: 4px;
  font-size: 13px;
  outline: none;
  background: #fff;
  color: #333;
}
.settings-row select:focus {
  border-color: var(--primary, #2d8cf0);
}

.section-block {
  background: rgba(0, 0, 0, 0.02);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 16px;
}

.segment-tabs {
  display: flex;
  background: rgba(0, 0, 0, 0.04);
  border-radius: 6px;
  padding: 3px;
  gap: 2px;
}
.seg-tab {
  flex: 1;
  padding: 6px 0;
  border: none;
  background: none;
  cursor: pointer;
  font-size: 13px;
  color: #666;
  border-radius: 4px;
  transition: all 0.15s;
}
.seg-tab.active {
  background: #fff;
  color: #333;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.color-presets-grid {
  display: grid;
  grid-template-columns: repeat(10, 1fr);
  gap: 6px;
  justify-items: center;
}
.color-dot {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  transition: all 0.15s;
  padding: 0;
  flex-shrink: 0;
}
.color-dot:hover { transform: scale(1.15); }
.color-dot.active {
  border-color: #333;
  box-shadow: 0 0 0 2px #fff, 0 0 0 4px currentColor;
}
.color-custom {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  overflow: hidden;
  border: 1px dashed #ccc;
  cursor: pointer;
  flex-shrink: 0;
}
.color-custom input {
  width: 100%;
  height: 100%;
  border: none;
  cursor: pointer;
  opacity: 0;
}

.radio-group {
  display: flex;
  gap: 4px;
}
.radio-btn {
  padding: 4px 12px;
  border: 0.8px solid #ddd;
  border-radius: 4px;
  font-size: 13px;
  cursor: pointer;
  color: #666;
  transition: all 0.15s;
  background: #fff;
}
.radio-btn:hover { border-color: #aaa; }
.radio-btn.active {
  border-color: var(--primary, #2d8cf0);
  color: var(--primary, #2d8cf0);
  background: rgba(45, 140, 240, 0.05);
}

.toggle {
  width: 40px;
  height: 22px;
  border-radius: 11px;
  background: #ddd;
  border: none;
  cursor: pointer;
  position: relative;
  transition: background 0.2s;
  padding: 0;
}
.toggle::after {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: #fff;
  transition: transform 0.2s;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
}
.toggle.on {
  background: var(--primary, #2d8cf0);
}
.toggle.on::after {
  transform: translateX(18px);
}
.toggle:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.num-group {
  display: flex;
  align-items: center;
  gap: 4px;
}
.num-btn {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 0.8px solid #ddd;
  border-radius: 4px;
  background: #fff;
  cursor: pointer;
  font-size: 14px;
  color: #666;
  transition: all 0.15s;
}
.num-btn:hover { border-color: var(--primary, #2d8cf0); color: var(--primary, #2d8cf0); }
.num-input {
  width: 48px;
  height: 24px;
  text-align: center;
  border: 0.8px solid #ddd;
  border-radius: 4px;
  font-size: 13px;
  outline: none;
  color: #333;
}
.num-input:focus { border-color: var(--primary, #2d8cf0); }

.layout-cards {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 8px;
}
.layout-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 8px 4px;
  border: 0.8px solid #eee;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s;
}
.layout-card:hover { border-color: #ccc; }
.layout-card.active {
  border-color: var(--primary, #2d8cf0);
  background: rgba(45, 140, 240, 0.04);
}
.layout-card span {
  font-size: 11px;
  color: #666;
}
.layout-card.active span { color: var(--primary, #2d8cf0); }
.layout-preview {
  display: flex;
  align-items: center;
  justify-content: center;
}

.pure-color-section {
  display: flex;
  justify-content: center;
  margin-bottom: 12px;
}

.skin-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 8px;
}
.skin-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 8px 4px;
  border: 0.8px solid #eee;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s;
}
.skin-card:hover { border-color: #ccc; }
.skin-card.active {
  border-color: var(--primary, #2d8cf0);
  background: rgba(45, 140, 240, 0.04);
}
.skin-card span {
  font-size: 11px;
  color: #666;
}
.skin-card.active span { color: var(--primary, #2d8cf0); }
.skin-preview {
  width: 100%;
  height: 24px;
  border-radius: 4px;
}

.settings-footer {
  padding: 12px 20px;
  border-top: 0.8px solid #eee;
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}
.btn {
  flex: 1;
  padding: 8px 0;
  border: 0.8px solid #ddd;
  border-radius: 6px;
  background: #fff;
  cursor: pointer;
  font-size: 13px;
  color: #666;
  transition: all 0.15s;
}
.btn:hover { border-color: #aaa; color: #333; }
.btn-primary {
  background: var(--primary, #2d8cf0);
  border-color: var(--primary, #2d8cf0);
  color: #fff;
}
.btn-primary:hover {
  background: var(--primary-hover, #1c6ac9);
  border-color: var(--primary-hover, #1c6ac9);
  color: #fff;
}
</style>
