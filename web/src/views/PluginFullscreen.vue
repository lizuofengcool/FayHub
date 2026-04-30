<template>
  <div class="plugin-fullscreen-page">
    <div class="fullscreen-header">
      <div class="flex items-center gap-3">
        <el-button text @click="goBack">
          <el-icon class="text-lg"><ArrowLeft /></el-icon>
          返回后台
        </el-button>
        <div class="h-5 w-px bg-slate-200"></div>
        <h2 class="text-lg font-bold text-slate-800">{{ pageTitle }}</h2>
        <span class="text-xs text-slate-400">由插件 {{ pluginId }} 提供</span>
        <el-tag v-if="renderMode === 'custom'" type="warning" size="small">自定义组件</el-tag>
        <el-tag v-else type="info" size="small">Schema驱动</el-tag>
      </div>
      <div class="flex gap-3">
        <el-button @click="handleRefresh" :loading="refreshing">
          <el-icon class="mr-1"><Refresh /></el-icon> 刷新
        </el-button>
      </div>
    </div>

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
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Refresh, ArrowLeft } from '@element-plus/icons-vue'
import HostLoader from '@/views/HostLoader.vue'
import request from '@/api/request'

const route = useRoute()
const router = useRouter()

const pluginId = computed(() => {
  const pluginPath = route.params.pluginPath as string
  return pluginPath || ''
})

const pageTitle = computed(() => {
  const path = route.params.pluginPath as string
  if (path) {
    const parts = path.split('/')
    return parts[parts.length - 1] || '插件应用'
  }
  return '插件应用'
})

const renderMode = ref<'custom' | 'schema'>('schema')
const useShadowDom = ref(false)
const manifest = ref<Record<string, any>>({})
const refreshing = ref(false)

function goBack() {
  router.push('/dashboard')
}

async function fetchPluginManifest() {
  if (!pluginId.value) return
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

onMounted(() => {
  fetchPluginManifest()
})

watch(() => route.params.pluginPath, () => {
  fetchPluginManifest()
})
</script>

<style scoped>
.plugin-fullscreen-page {
  min-height: 100vh;
  background: #f8fafc;
  padding: 24px;
}
.fullscreen-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
  padding: 16px 24px;
  background: white;
  border-radius: 16px;
  border: 1px solid rgba(226, 232, 240, 0.8);
}
</style>
