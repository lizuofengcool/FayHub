<template>
  <div class="plugin-devtools">
    <div class="devtools-header">
      <h3>FayHub Plugin DevTools</h3>
      <el-button size="small" @click="refresh">刷新</el-button>
    </div>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="已加载插件" name="plugins">
        <el-table :data="plugins" size="small" stripe>
          <el-table-column prop="pluginId" label="ID" width="180" />
          <el-table-column prop="manifest.name" label="名称" width="150" />
          <el-table-column prop="manifest.version" label="版本" width="80" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="manifest.renderMode" label="渲染模式" width="100" />
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button size="small" type="danger" @click="handleUnload(row.pluginId)">卸载</el-button>
              <el-button size="small" @click="handleReload(row.pluginId)">重载</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="沙箱状态" name="sandbox">
        <div v-for="p in plugins" :key="p.pluginId" class="sandbox-info">
          <h4>{{ p.pluginId }}</h4>
          <div class="info-grid">
            <span>沙箱已销毁:</span>
            <span>{{ p.sandbox?.isDestroyed?.() ?? 'N/A' }}</span>
            <span>Bridge 插件ID:</span>
            <span>{{ p.bridge?.getPluginId?.() ?? 'N/A' }}</span>
            <span>Bridge 租户ID:</span>
            <span>{{ p.bridge?.getTenantId?.() ?? 'N/A' }}</span>
          </div>
        </div>
        <el-empty v-if="plugins.length === 0" description="暂无已加载插件" />
      </el-tab-pane>

      <el-tab-pane label="事件日志" name="events">
        <div class="event-log">
          <div v-for="(log, i) in eventLogs" :key="i" class="log-entry" :class="log.level">
            <span class="log-time">{{ log.time }}</span>
            <span class="log-level">[{{ log.level }}]</span>
            <span class="log-msg">{{ log.message }}</span>
          </div>
        </div>
        <el-empty v-if="eventLogs.length === 0" description="暂无事件日志" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { getAllLoadedPlugins, unloadPlugin, getLoadedPlugin } from '@/plugin/loader'
import { loadPlugin } from '@/plugin/loader'

const activeTab = ref('plugins')
const plugins = ref([])
const eventLogs = ref([])
let refreshTimer = null

function refresh() {
  plugins.value = getAllLoadedPlugins().map(p => ({
    pluginId: p.pluginId,
    manifest: p.manifest,
    status: p.status,
    sandbox: p.sandbox,
    bridge: p.bridge,
    error: p.error,
  }))
}

function statusType(status) {
  if (status === 'loaded') return 'success'
  if (status === 'loading') return 'warning'
  if (status === 'error') return 'danger'
  return 'info'
}

function handleUnload(pluginId) {
  unloadPlugin(pluginId)
  addLog('info', `插件 ${pluginId} 已卸载`)
  refresh()
}

async function handleReload(pluginId) {
  const existing = getLoadedPlugin(pluginId)
  if (!existing) return

  unloadPlugin(pluginId)

  try {
    await loadPlugin(existing.manifest, existing.bridge?.getTenantId?.() || 0)
    addLog('info', `插件 ${pluginId} 重新加载成功`)
  } catch (err) {
    addLog('error', `插件 ${pluginId} 重新加载失败: ${err.message}`)
  }
  refresh()
}

function addLog(level, message) {
  eventLogs.value.unshift({
    time: new Date().toLocaleTimeString(),
    level,
    message,
  })
  if (eventLogs.value.length > 100) {
    eventLogs.value = eventLogs.value.slice(0, 100)
  }
}

onMounted(() => {
  refresh()
  refreshTimer = setInterval(refresh, 3000)
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<style scoped>
.plugin-devtools {
  position: fixed;
  bottom: 0;
  right: 0;
  width: 600px;
  max-height: 400px;
  background: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 8px 0 0 0;
  box-shadow: 0 -2px 12px rgba(0, 0, 0, 0.1);
  z-index: 99999;
  overflow: auto;
  font-size: 12px;
}
.devtools-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid #ebeef5;
  background: #f5f7fa;
}
.devtools-header h3 {
  margin: 0;
  font-size: 14px;
  color: #303133;
}
.sandbox-info {
  padding: 8px 12px;
  border-bottom: 1px solid #ebeef5;
}
.sandbox-info h4 {
  margin: 0 0 4px 0;
  font-size: 13px;
  color: #409eff;
}
.info-grid {
  display: grid;
  grid-template-columns: 120px 1fr;
  gap: 2px 8px;
  font-size: 12px;
}
.event-log {
  max-height: 250px;
  overflow-y: auto;
  padding: 4px;
}
.log-entry {
  padding: 2px 8px;
  font-family: monospace;
  font-size: 11px;
  border-bottom: 1px solid #f0f0f0;
}
.log-entry.error { color: #f56c6c; }
.log-entry.warn { color: #e6a23c; }
.log-entry.info { color: #909399; }
.log-time { color: #c0c4cc; margin-right: 4px; }
.log-level { margin-right: 4px; font-weight: bold; }
</style>
