<template>
  <div class="plugin-version-page">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-slate-800">插件版本管理</h2>
        <p class="text-slate-500 mt-1 text-sm">查看插件版本历史、依赖关系与更新</p>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div class="lg:col-span-1">
        <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
          <div class="p-4 border-b border-slate-100">
            <h3 class="font-semibold text-slate-700">已安装插件</h3>
          </div>
          <div v-loading="pluginLoading">
            <div
              v-for="plugin in plugins"
              :key="plugin.plugin_id"
              class="p-4 border-b border-slate-50 cursor-pointer hover:bg-slate-50 transition-colors"
              :class="{ 'bg-blue-50 border-l-4 border-l-blue-500': selectedPlugin?.plugin_id === plugin.plugin_id }"
              @click="selectPlugin(plugin)"
            >
              <div class="flex items-center justify-between">
                <div>
                  <p class="font-medium text-slate-800">{{ plugin.name }}</p>
                  <p class="text-xs text-slate-400">{{ plugin.plugin_id }}</p>
                </div>
                <el-tag :type="plugin.status === 'active' ? 'success' : 'warning'" size="small">
                  {{ plugin.status === 'active' ? '运行中' : '已禁用' }}
                </el-tag>
              </div>
              <div class="flex items-center gap-2 mt-2">
                <span class="text-xs text-slate-500">v{{ plugin.version }}</span>
                <el-tag v-if="updateInfo[plugin.plugin_id]?.has_update" type="danger" size="small" effect="dark">
                  有更新
                </el-tag>
              </div>
            </div>
            <div v-if="plugins.length === 0 && !pluginLoading" class="text-center py-8 text-slate-400 text-sm">
              暂无已安装插件
            </div>
          </div>
        </div>
      </div>

      <div class="lg:col-span-2">
        <div v-if="selectedPlugin">
          <el-tabs v-model="activeTab">
            <el-tab-pane label="版本历史" name="versions">
              <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
                <div v-loading="versionLoading">
                  <el-timeline class="p-6">
                    <el-timeline-item
                      v-for="ver in versions"
                      :key="ver.id"
                      :timestamp="ver.created_at"
                      placement="top"
                      :type="ver.is_latest ? 'primary' : 'info'"
                    >
                      <div class="bg-slate-50 rounded-lg p-4">
                        <div class="flex items-center justify-between mb-2">
                          <span class="font-semibold text-slate-800">v{{ ver.version }}</span>
                          <el-tag v-if="ver.is_latest" type="success" size="small">当前版本</el-tag>
                        </div>
                        <p class="text-sm text-slate-600">{{ ver.changelog || '无更新说明' }}</p>
                        <div class="flex gap-4 mt-2 text-xs text-slate-400">
                          <span v-if="ver.min_engine_version">最低引擎版本: {{ ver.min_engine_version }}</span>
                          <span v-if="ver.signature">签名: {{ ver.signature.substring(0, 16) }}...</span>
                        </div>
                      </div>
                    </el-timeline-item>
                  </el-timeline>
                  <div v-if="versions.length === 0 && !versionLoading" class="text-center py-8 text-slate-400 text-sm">
                    暂无版本记录
                  </div>
                </div>
              </div>
            </el-tab-pane>

            <el-tab-pane label="操作记录" name="history">
              <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
                <el-table v-loading="historyLoading" :data="history" stripe class="w-full">
                  <el-table-column prop="action" label="操作" width="120">
                    <template #default="{ row }">
                      <el-tag :type="actionTagType(row.action)" size="small">{{ actionLabel(row.action) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column label="版本变更" min-width="180">
                    <template #default="{ row }">
                      <span v-if="row.from_version" class="text-slate-500">v{{ row.from_version }}</span>
                      <span v-if="row.from_version && row.to_version" class="mx-2 text-slate-400">→</span>
                      <span v-if="row.to_version" class="font-medium text-slate-800">v{{ row.to_version }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column prop="operator" label="操作人" width="120" />
                  <el-table-column prop="created_at" label="时间" width="160" />
                </el-table>
                <div class="p-4 flex justify-end">
                  <el-pagination
                    v-model:current-page="historyPage"
                    :page-size="10"
                    :total="historyTotal"
                    layout="total, prev, pager, next"
                    @current-change="fetchHistory"
                  />
                </div>
              </div>
            </el-tab-pane>

            <el-tab-pane label="依赖关系" name="dependencies">
              <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
                <el-table v-loading="depLoading" :data="dependencies" stripe class="w-full">
                  <el-table-column prop="dependency_plugin_id" label="依赖插件" min-width="200" />
                  <el-table-column prop="dependency_version" label="依赖版本" width="140" />
                  <el-table-column prop="is_required" label="必需" width="80" align="center">
                    <template #default="{ row }">
                      <el-tag :type="row.is_required ? 'danger' : 'info'" size="small">
                        {{ row.is_required ? '必需' : '可选' }}
                      </el-tag>
                    </template>
                  </el-table-column>
                </el-table>
                <div v-if="dependencies.length === 0 && !depLoading" class="text-center py-8 text-slate-400 text-sm">
                  无依赖关系
                </div>
              </div>
            </el-tab-pane>
          </el-tabs>

          <div v-if="updateInfo[selectedPlugin.plugin_id]?.has_update" class="mt-4 bg-blue-50 border border-blue-200 rounded-xl p-4">
            <div class="flex items-center justify-between">
              <div>
                <p class="font-medium text-blue-800">发现新版本 v{{ updateInfo[selectedPlugin.plugin_id].latest_version }}</p>
                <p class="text-sm text-blue-600 mt-1">{{ updateInfo[selectedPlugin.plugin_id].changelog || '暂无更新说明' }}</p>
              </div>
              <el-button type="primary" size="default" @click="handleUpgrade">立即升级</el-button>
            </div>
          </div>
        </div>

        <div v-else class="bg-white rounded-2xl border border-slate-100 shadow-sm flex items-center justify-center" style="min-height: 400px;">
          <div class="text-center text-slate-400">
            <el-icon class="text-5xl mb-3"><Box /></el-icon>
            <p>请从左侧选择一个插件查看详情</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Box } from '@element-plus/icons-vue'
import pluginEngineApi, { type InstalledPlugin } from '@/api/pluginEngine'
import pluginVersionApi, { type PluginVersion, type PluginVersionHistory, type PluginDependency } from '@/api/pluginVersion'

const pluginLoading = ref(false)
const plugins = ref<InstalledPlugin[]>([])
const selectedPlugin = ref<InstalledPlugin | null>(null)
const updateInfo = reactive<Record<string, { has_update: boolean; latest_version: string; changelog: string }>>({})

const activeTab = ref('versions')
const versionLoading = ref(false)
const versions = ref<PluginVersion[]>([])

const historyLoading = ref(false)
const history = ref<PluginVersionHistory[]>([])
const historyPage = ref(1)
const historyTotal = ref(0)

const depLoading = ref(false)
const dependencies = ref<PluginDependency[]>([])

async function fetchPlugins() {
  pluginLoading.value = true
  try {
    const res = await pluginEngineApi.getInstalledPlugins()
    const data = res.data
    if (Array.isArray(data)) {
      plugins.value = data
    } else if (data?.list) {
      plugins.value = data.list
    } else {
      plugins.value = []
    }
    plugins.value.forEach(p => checkUpdate(p.plugin_id))
  } catch (err: any) {
    ElMessage.error(err.message || '获取插件列表失败')
  } finally {
    pluginLoading.value = false
  }
}

async function checkUpdate(pluginId: string) {
  try {
    const res = await pluginVersionApi.checkUpdates(pluginId)
    if (res.data) {
      updateInfo[pluginId] = res.data
    }
  } catch (e) { console.error('checkUpdates failed:', e); }
}

function selectPlugin(plugin: InstalledPlugin) {
  selectedPlugin.value = plugin
  activeTab.value = 'versions'
  fetchVersions()
  fetchHistory()
  fetchDependencies()
}

async function fetchVersions() {
  if (!selectedPlugin.value) return
  versionLoading.value = true
  try {
    const res = await pluginVersionApi.listVersions(selectedPlugin.value.plugin_id)
    versions.value = res.data || []
  } catch (err: any) {
    ElMessage.error(err.message || '获取版本列表失败')
  } finally {
    versionLoading.value = false
  }
}

async function fetchHistory() {
  if (!selectedPlugin.value) return
  historyLoading.value = true
  try {
    const res = await pluginVersionApi.listVersionHistory(selectedPlugin.value.plugin_id, {
      page: historyPage.value,
      page_size: 10
    })
    history.value = res.data?.list || []
    historyTotal.value = res.data?.total || 0
  } catch (err: any) {
    ElMessage.error(err.message || '获取操作记录失败')
  } finally {
    historyLoading.value = false
  }
}

async function fetchDependencies() {
  if (!selectedPlugin.value) return
  depLoading.value = true
  try {
    const res = await pluginVersionApi.listDependencies(selectedPlugin.value.plugin_id)
    dependencies.value = res.data || []
  } catch (err: any) {
    ElMessage.error(err.message || '获取依赖关系失败')
  } finally {
    depLoading.value = false
  }
}

async function handleUpgrade() {
  if (!selectedPlugin.value) return
  const info = updateInfo[selectedPlugin.value.plugin_id]
  if (!info) return

  try {
    await ElMessageBox.confirm(
      `确定要将插件从 v${selectedPlugin.value.version} 升级到 v${info.latest_version} 吗？`,
      '确认升级',
      { type: 'warning' }
    )
    await pluginEngineApi.upgradePlugin(selectedPlugin.value.plugin_id, info.latest_version, '')
    ElMessage.success('升级成功')
    fetchPlugins()
    selectedPlugin.value = null
  } catch (e) { console.error('handleUpgrade failed:', e); }
}

function actionTagType(action: string) {
  switch (action) {
    case 'install': return 'success'
    case 'uninstall': return 'danger'
    case 'upgrade': return 'primary'
    case 'rollback': return 'warning'
    case 'enable': return 'success'
    case 'disable': return 'info'
    default: return 'info'
  }
}

function actionLabel(action: string) {
  switch (action) {
    case 'install': return '安装'
    case 'uninstall': return '卸载'
    case 'upgrade': return '升级'
    case 'rollback': return '回滚'
    case 'enable': return '启用'
    case 'disable': return '禁用'
    default: return action
  }
}

onMounted(() => {
  fetchPlugins()
})
</script>

<style scoped>
</style>
