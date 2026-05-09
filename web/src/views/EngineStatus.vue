﻿<template>
  <div class="engine-status-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">插件引擎</h2>
          <p class="text-slate-400 text-xs mt-0.5">管理插件运行时引擎状态与已加载插件</p>
        </div>
        <el-button @click="refreshAll" :loading="loading">
          <el-icon class="mr-1"><Refresh /></el-icon> 刷新
        </el-button>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 px-4 mb-4">
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
        <div class="flex items-center justify-between mb-4">
          <span class="text-sm font-medium text-slate-500">引擎类型</span>
          <el-icon class="text-2xl" :class="status.is_running ? 'text-green-500' : 'text-red-400'">
            <Cpu />
          </el-icon>
        </div>
        <p class="text-3xl font-bold text-slate-800">{{ status.engine_type === 'wasm' ? 'WASM' : 'Noop' }}</p>
        <p class="text-sm mt-1" :class="status.is_running ? 'text-green-600' : 'text-red-500'">
          {{ status.is_running ? '运行中' : '未启动' }}
        </p>
      </div>

      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
        <div class="flex items-center justify-between mb-4">
          <span class="text-sm font-medium text-slate-500">已加载插件</span>
          <el-icon class="text-2xl text-blue-500"><Box /></el-icon>
        </div>
        <p class="text-3xl font-bold text-slate-800">{{ status.plugin_count || 0 }}</p>
        <p class="text-sm text-slate-500 mt-1">当前运行时实例</p>
      </div>

      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
        <div class="flex items-center justify-between mb-4">
          <span class="text-sm font-medium text-slate-500">运行时状态</span>
          <el-icon class="text-2xl text-indigo-500"><CircleCheck /></el-icon>
        </div>
        <p class="text-3xl font-bold" :class="status.is_running ? 'text-green-600' : 'text-red-500'">
          {{ status.is_running ? '正常' : '异常' }}
        </p>
        <p class="text-sm text-slate-500 mt-1">基于 wazero 纯Go运行时</p>
      </div>
    </div>

    <div class="bg-white rounded-2xl shadow-sm border border-slate-100">
      <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
        <h3 class="text-lg font-semibold text-slate-800">已加载插件列表</h3>
        <n-tag :type="plugins.length > 0 ? 'success' : 'info'" size="small">
          {{ plugins.length }} 个插件
        </n-tag>
      </div>

      <el-table :data="plugins" stripe class="w-full" empty-text="暂无已加载插件">
        <el-table-column prop="key" label="标识" min-width="120" />
        <el-table-column prop="name" label="插件名称" min-width="150" />
        <el-table-column prop="version" label="版本" width="100" />
        <el-table-column prop="entry_point" label="入口函数" width="120" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <n-tag :type="row.status === 'active' ? 'success' : 'warning'" size="small">
              {{ row.status === 'active' ? '运行中' : '已禁用' }}
            </n-tag>
          </template>
        </el-table-column>
        <el-table-column label="WASM模块" width="100">
          <template #default="{ row }">
            <n-tag :type="row.has_module ? 'success' : 'info'" size="small">
              {{ row.has_module ? '已加载' : '仅元数据' }}
            </n-tag>
          </template>
        </el-table-column>
        <el-table-column label="权限" min-width="200">
          <template #default="{ row }">
            <div class="flex flex-wrap gap-1">
              <n-tag v-for="perm in (row.permissions || [])" :key="perm" size="small" type="default" class="text-xs">
                {{ perm }}
              </n-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="checkHealth(row)">健康检查</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getEngineStatus, getLoadedPlugins, healthCheckPlugin } from '@/api/engine'
import { useMessage } from 'naive-ui'
const message = useMessage()
import { Refresh, Cpu, Box, CircleCheck } from '@element-plus/icons-vue'

interface EngineStatus {
  engine_type: string
  is_running: boolean
  plugin_count: number
}

interface PluginInfo {
	key: string
	plugin_id: string
	name: string
	version: string
	status: string
	entry_point: string
	permissions: string[]
	has_module: boolean
}

const loading = ref(false)
const status = ref<EngineStatus>({
  engine_type: 'noop',
  is_running: false,
  plugin_count: 0
})
const plugins = ref<PluginInfo[]>([])

async function refreshAll() {
  loading.value = true
  try {
    const [statusRes, pluginsRes]: any[] = await Promise.all([
      getEngineStatus(),
      getLoadedPlugins()
    ])
    if (statusRes.data) {
      status.value = statusRes.data
    }
    if (pluginsRes.data) {
      plugins.value = pluginsRes.data
    }
  } catch (e: any) {
    message.error('获取引擎状态失败: ' + (e.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

async function checkHealth(plugin: PluginInfo) {
	try {
		if (!plugin.plugin_id) {
			message.warning('无法识别插件ID')
			return
		}
		const res: any = await healthCheckPlugin(plugin.plugin_id)
    if (res.data && res.data.status === 'healthy') {
      message.success(`插件 ${plugin.name} 运行正常`)
    } else {
      message.warning(`插件 ${plugin.name} 状态异常`)
    }
  } catch (e: any) {
    message.error(`插件 ${plugin.name} 健康检查失败: ` + (e.message || '未知错误'))
  }
}

onMounted(() => {
  refreshAll()
})
</script>

<style scoped>
</style>
