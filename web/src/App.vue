<template>
  <n-config-provider :theme="naiveTheme" :theme-overrides="themeOverrides" :locale="zhCN" :date-locale="dateZhCN">
    <n-message-provider>
      <n-dialog-provider>
        <n-notification-provider>
          <div id="app">
            <div class="bg-mesh">
              <div class="blob blob-1"></div>
              <div class="blob blob-2"></div>
              <div class="blob blob-3"></div>
            </div>

            <div class="app-content">
              <router-view />
            </div>
          </div>
        </n-notification-provider>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { darkTheme, zhCN, dateZhCN, type GlobalThemeOverrides, NConfigProvider, NMessageProvider, NDialogProvider, NNotificationProvider } from 'naive-ui'
import { useThemeStore } from '@/stores/theme'
import { usePreferencesStore } from '@/stores/preferences'

const themeStore = useThemeStore()
const preferencesStore = usePreferencesStore()

const naiveTheme = computed(() => themeStore.isDark ? darkTheme : null)

function lightenColor(hex: string, amount: number): string {
  const num = parseInt(hex.replace('#', ''), 16)
  const r = Math.min(255, (num >> 16) + Math.round(255 * amount))
  const g = Math.min(255, ((num >> 8) & 0x00FF) + Math.round(255 * amount))
  const b = Math.min(255, (num & 0x0000FF) + Math.round(255 * amount))
  return '#' + (0x1000000 + (r << 16) + (g << 8) + b).toString(16).slice(1)
}

function darkenColor(hex: string, amount: number): string {
  const num = parseInt(hex.replace('#', ''), 16)
  const r = Math.max(0, (num >> 16) - Math.round(255 * amount))
  const g = Math.max(0, ((num >> 8) & 0x00FF) - Math.round(255 * amount))
  const b = Math.max(0, (num & 0x0000FF) - Math.round(255 * amount))
  return '#' + (0x1000000 + (r << 16) + (g << 8) + b).toString(16).slice(1)
}

const themeOverrides = computed<GlobalThemeOverrides>(() => {
  const c = preferencesStore.prefs.themeColor || '#4f46e5'
  return {
    common: {
      primaryColor: c,
      primaryColorHover: lightenColor(c, 0.1),
      primaryColorPressed: darkenColor(c, 0.05),
      primaryColorSuppl: c,
      borderRadius: preferencesStore.prefs.radius ? parseFloat(preferencesStore.prefs.radius) * 16 + 'px' : '8px',
    },
  }
})
</script>

<style scoped>
.bg-mesh {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  z-index: 0;
  overflow: hidden;
  background: #f1f5f9;
}

.app-content {
  position: relative;
  z-index: 1;
}

.blob {
  position: absolute;
  filter: blur(80px);
  z-index: -1;
  opacity: 0.3;
  animation: float 20s infinite ease-in-out alternate;
}

.blob-1 {
  top: -20%; left: -10%;
  width: 40vw; height: 40vw;
  background: radial-gradient(circle, rgba(99,102,241,0.2) 0%, rgba(99,102,241,0) 70%);
  animation-delay: 0s;
}

.blob-2 {
  bottom: -30%; right: -10%;
  width: 50vw; height: 50vw;
  background: radial-gradient(circle, rgba(14,165,233,0.15) 0%, rgba(14,165,233,0) 70%);
  animation-delay: -5s;
}

.blob-3 {
  top: 20%; left: 70%;
  width: 35vw; height: 35vw;
  background: radial-gradient(circle, rgba(168,85,247,0.15) 0%, rgba(168,85,247,0) 70%);
  animation-delay: -10s;
}

@keyframes float {
  0% { transform: translate(0, 0) scale(1); }
  50% { transform: translate(5%, 10%) scale(1.1); }
  100% { transform: translate(-5%, 5%) scale(0.9); }
}
</style>