<template>
  <div class="plugin-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">{{ pageTitle }}</h2>
          <p class="text-slate-400 text-xs mt-0.5">
            由插件 {{ pluginId }} 提供
            <el-tag v-if="renderMode === 'custom'" type="warning" size="small" class="ml-2">自定义组件</el-tag>
            <el-tag v-else type="info" size="small" class="ml-2">Schema驱动</el-tag>
          </p>
        </div>
        <div class="flex gap-3">
          <el-button @click="handleRefresh" :loading="refreshing">
            <el-icon class="mr-1"><Refresh /></el-icon> 刷新
          </el-button>
        </div>
      </div>

      <div class="p-6">
        <HostLoader
          :plugin-id="pluginId"
          :render-mode="renderMode"
          :use-shadow-dom="useShadowDom"
          :manifest="manifest"
          @loaded="onPluginLoaded"
          @error="onPluginError"
          @unloaded="onPluginUnloaded"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import HostLoader from '@/views/HostLoader.vue'
import request from '@/api/request'

const route = useRoute()

const pluginId = ref('')
const renderMode = ref<'custom' | 'schema'>('schema')
const useShadowDom = ref(false)
const manifest = ref<Record<string, any>>({})
const refreshing = ref(false)

const pageTitle = computed(() => (route.meta.title as string) || '插件页面')

const menuTreeData = ref<any[]>([])

async function loadMenuTreeForLookup() {
  if (menuTreeData.value.length > 0) return
  try {
    const res = await request.get('/menus/tree')
    menuTreeData.value = res.data || []
  } catch (e) { console.error('fetchMenuTree failed:', e); }
}

async function resolvePluginInfo() {
  if (route.meta.pluginId) {
    pluginId.value = route.meta.pluginId as string
  } else {
    await loadMenuTreeForLookup()
    const currentPath = route.path
    for (const menu of menuTreeData.value) {
      if (menu.children) {
        for (const child of menu.children) {
          if (child.path === currentPath && child.component) {
            pluginId.value = child.component
            break
          }
        }
      }
    }
  }

  if (pluginId.value) {
    await fetchPluginManifest()
  } else {
    ElMessage.warning('该插件尚未安装或已被卸载')
  }
}

async function fetchPluginManifest() {
  try {
    const res = await request.get(`/plugin-engine/plugins/${pluginId.value}`)
    const data = res.data || {}

    renderMode.value = data.render_mode || data.renderMode || 'schema'
    useShadowDom.value = data.use_shadow_dom || data.useShadowDom || false

    if (renderMode.value === 'custom') {
      manifest.value = {
        name: data.name || pluginId.value,
        version: data.version || '1.0.0',
        entry: data.entry || 'index.js',
        style: data.style,
        permissions: data.permissions,
        allowedApiPrefixes: data.allowed_api_prefixes || data.allowedApiPrefixes,
        compatibleBaseVersion: data.compatible_base_version || data.compatibleBaseVersion,
        signature: data.signature,
      }
    }
  } catch {
    renderMode.value = 'schema'
  }
}

function handleRefresh() {
  refreshing.value = true
  window.location.reload()
}

function onPluginLoaded(id: string) {
  refreshing.value = false
}

function onPluginError(id: string, error: string) {
  refreshing.value = false
}

function onPluginUnloaded(id: string) {}

onMounted(async () => {
  await resolvePluginInfo()
})

watch(() => route.path, async () => {
  await resolvePluginInfo()
})
</script>

<style scoped>
.plugin-page {
  min-height: 100%;
}
</style>
