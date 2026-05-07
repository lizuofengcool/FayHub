<template>
  <Teleport to="body">
    <Transition name="search-fade">
      <div v-if="visible" class="search-overlay" @click="$emit('close')"></div>
    </Transition>
    <Transition name="search-slide">
      <div v-if="visible" class="search-dialog">
        <div class="search-input-wrap">
          <svg class="search-icon" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
          </svg>
          <input
            ref="inputRef"
            v-model="query"
            type="text"
            placeholder="输入内容、支持按首字母搜索"
            @keydown="handleKeydown"
          >
          <kbd class="search-kbd">Esc</kbd>
        </div>

        <div class="search-body">
          <div class="search-results" v-if="results.length > 0">
            <div class="search-result-group" v-for="group in groupedResults" :key="group.label">
              <div class="search-group-label">{{ group.label }}</div>
              <div
                v-for="(item, idx) in group.items"
                :key="item.path"
                class="search-result-item"
                :class="{ active: activeIndex === getGlobalIndex(group, idx) }"
                @click="navigateTo(item)"
                @mouseenter="activeIndex = getGlobalIndex(group, idx)"
              >
                <span class="result-icon">{{ item.icon }}</span>
                <div class="result-info">
                  <div class="result-title">{{ item.title }}</div>
                  <div class="result-path">{{ item.path }}</div>
                </div>
                <span class="result-enter">↵</span>
              </div>
            </div>
          </div>

          <div class="search-empty" v-else-if="query.length > 0">
            <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.3">
              <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
            </svg>
            <p>未找到相关结果</p>
          </div>

          <div class="search-initial" v-else>
            <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.3">
              <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
            </svg>
            <p>输入搜索关键字开始搜索吧~</p>
          </div>
        </div>

        <div class="search-footer">
          <div class="footer-hints">
            <span><kbd>↑↓</kbd> 选择</span>
            <span><kbd>Enter</kbd> 确认</span>
            <span><kbd>Esc</kbd> 关闭</span>
          </div>
          <div class="footer-actions">
            <button class="footer-btn" @click="handleEnter">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
              选择
            </button>
            <button class="footer-btn" @click="handleNewWindow">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6M15 3h6v6M10 14 21 3"/></svg>
              新窗口
            </button>
            <button class="footer-btn" @click="$emit('close')">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 6 6 18M6 6l12 12"/></svg>
              退出
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ close: [] }>()

const router = useRouter()
const query = ref('')
const activeIndex = ref(0)
const inputRef = ref<HTMLInputElement | null>(null)

interface SearchItem {
  title: string
  path: string
  icon: string
  group: string
}

const searchItems: SearchItem[] = [
  { title: '主控台', path: '/dashboard', icon: '📊', group: '仪表盘' },
  { title: '监控页', path: '/monitor', icon: '📈', group: '仪表盘' },
  { title: '工作台', path: '/workbench', icon: '💼', group: '仪表盘' },
  { title: '租户管理', path: '/system/tenant', icon: '🏢', group: '系统管理' },
  { title: '用户管理', path: '/system/user', icon: '👥', group: '系统管理' },
  { title: '角色管理', path: '/system/role', icon: '🔑', group: '系统管理' },
  { title: '部门管理', path: '/system/department', icon: '🏛', group: '系统管理' },
  { title: '菜单管理', path: '/system/menu', icon: '📋', group: '系统管理' },
  { title: '字典管理', path: '/system/dict', icon: '📖', group: '系统管理' },
  { title: '通知管理', path: '/system/notifications', icon: '🔔', group: '系统管理' },
  { title: '审计日志', path: '/system/audit', icon: '📝', group: '系统管理' },
  { title: '登录日志', path: '/system/login-logs', icon: '🔐', group: '系统管理' },
  { title: '错误码', path: '/system/error-codes', icon: '❌', group: '系统管理' },
  { title: '在线用户', path: '/system/online-users', icon: '🟢', group: '系统管理' },
  { title: '定时任务', path: '/system/cron-jobs', icon: '⏰', group: '系统管理' },
  { title: '订阅管理', path: '/system/subscriptions', icon: '📬', group: '系统管理' },
  { title: '通知渠道', path: '/system/notification-channels', icon: '📡', group: '系统管理' },
  { title: '文件管理', path: '/system/files', icon: '📁', group: '系统管理' },
  { title: 'Webhook', path: '/system/webhook', icon: '🔗', group: '系统管理' },
  { title: '已安装插件', path: '/plugins/installed', icon: '🧩', group: '插件市场' },
  { title: '交易记录', path: '/payment/transactions', icon: '💳', group: '支付' },
  { title: '个人中心', path: '/profile', icon: '👤', group: '个人' },
]

const results = computed(() => {
  if (!query.value.trim()) return []
  const q = query.value.toLowerCase()
  return searchItems.filter(
    item =>
      item.title.toLowerCase().includes(q) ||
      item.path.toLowerCase().includes(q) ||
      item.group.toLowerCase().includes(q)
  )
})

const groupedResults = computed(() => {
  const groups: Record<string, SearchItem[]> = {}
  for (const item of results.value) {
    if (!groups[item.group]) groups[item.group] = []
    groups[item.group].push(item)
  }
  return Object.entries(groups).map(([label, items]) => ({ label, items }))
})

function getGlobalIndex(group: { label: string; items: SearchItem[] }, idx: number): number {
  let count = 0
  for (const g of groupedResults.value) {
    if (g.label === group.label) return count + idx
    count += g.items.length
  }
  return 0
}

function getItemByGlobalIndex(index: number): SearchItem | null {
  let count = 0
  for (const g of groupedResults.value) {
    if (index < count + g.items.length) {
      return g.items[index - count]
    }
    count += g.items.length
  }
  return null
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    const total = results.value.length
    if (total > 0) activeIndex.value = (activeIndex.value + 1) % total
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    const total = results.value.length
    if (total > 0) activeIndex.value = (activeIndex.value - 1 + total) % total
  } else if (e.key === 'Enter') {
    e.preventDefault()
    handleEnter()
  } else if (e.key === 'Escape') {
    emit('close')
  }
}

function handleEnter() {
  const item = getItemByGlobalIndex(activeIndex.value)
  if (item) navigateTo(item)
}

function handleNewWindow() {
  const item = getItemByGlobalIndex(activeIndex.value)
  if (item) {
    window.open(item.path, '_blank')
    emit('close')
  }
}

function navigateTo(item: SearchItem) {
  router.push(item.path)
  emit('close')
}

watch(() => props.visible, (val) => {
  if (val) {
    query.value = ''
    activeIndex.value = 0
    nextTick(() => {
      inputRef.value?.focus()
    })
  }
})
</script>

<style scoped>
.search-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  z-index: 1100;
}

.search-dialog {
  position: fixed;
  top: 18%;
  left: 50%;
  transform: translateX(-50%);
  width: 600px;
  max-height: 75vh;
  background: #fff;
  border-radius: 14px;
  box-shadow: 0 12px 48px rgba(0, 0, 0, 0.18);
  z-index: 1101;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.search-input-wrap {
  display: flex;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color, #eee);
  gap: 12px;
}
.search-icon {
  color: var(--text-muted, #999);
  flex-shrink: 0;
}
.search-input-wrap input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 16px;
  color: var(--text-primary, #333);
  background: transparent;
}
.search-input-wrap input::placeholder {
  color: var(--text-muted, #bbb);
}
.search-kbd {
  padding: 3px 10px;
  background: rgba(0, 0, 0, 0.06);
  border-radius: 5px;
  font-size: 12px;
  color: var(--text-muted, #999);
  font-family: inherit;
  font-weight: 500;
}

.search-body {
  overflow-y: auto;
  max-height: 45vh;
  min-height: 80px;
}

.search-results {
  padding: 6px;
}

.search-result-group {
  margin-bottom: 2px;
}
.search-group-label {
  padding: 8px 14px 4px;
  font-size: 11px;
  color: var(--text-muted, #999);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.8px;
}

.search-result-item {
  display: flex;
  align-items: center;
  padding: 10px 14px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.1s;
  gap: 12px;
}
.search-result-item:hover,
.search-result-item.active {
  background: rgba(45, 140, 240, 0.07);
}
.result-icon {
  font-size: 20px;
  width: 28px;
  text-align: center;
  flex-shrink: 0;
}
.result-info {
  flex: 1;
  min-width: 0;
}
.result-title {
  font-size: 14px;
  color: var(--text-primary, #333);
  font-weight: 500;
}
.result-path {
  font-size: 12px;
  color: var(--text-muted, #999);
  margin-top: 2px;
}
.result-enter {
  font-size: 14px;
  color: var(--text-muted, #ccc);
  flex-shrink: 0;
}

.search-empty,
.search-initial {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  gap: 12px;
  color: var(--text-muted, #999);
  font-size: 14px;
}
.search-empty p,
.search-initial p {
  margin: 0;
}

.search-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  border-top: 1px solid var(--border-color, #eee);
  background: rgba(0, 0, 0, 0.01);
}

.footer-hints {
  display: flex;
  gap: 14px;
  font-size: 12px;
  color: var(--text-muted, #999);
}
.footer-hints kbd {
  padding: 2px 6px;
  background: rgba(0, 0, 0, 0.06);
  border-radius: 4px;
  font-size: 11px;
  font-family: inherit;
}

.footer-actions {
  display: flex;
  gap: 6px;
}
.footer-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 5px 12px;
  border: 1px solid var(--border-color, #e0e0e0);
  border-radius: 6px;
  background: #fff;
  color: var(--text-secondary, #666);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}
.footer-btn:hover {
  border-color: var(--primary, #2d8cf0);
  color: var(--primary, #2d8cf0);
  background: rgba(45, 140, 240, 0.04);
}

.search-fade-enter-active,
.search-fade-leave-active {
  transition: opacity 0.2s;
}
.search-fade-enter-from,
.search-fade-leave-to {
  opacity: 0;
}

.search-slide-enter-active {
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}
.search-slide-leave-active {
  transition: all 0.15s cubic-bezier(0.4, 0, 0.2, 1);
}
.search-slide-enter-from {
  opacity: 0;
  transform: translateX(-50%) translateY(-10px) scale(0.97);
}
.search-slide-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(-4px) scale(0.98);
}
</style>
