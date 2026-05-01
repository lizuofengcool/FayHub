<template>
  <div class="plugins-page">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-slate-800">插件管理</h2>
        <p class="text-slate-500 mt-1 text-sm">管理已安装的插件，开启/禁用插件</p>
      </div>
      <el-button type="primary" @click="handleInstallDemo">
        <el-icon class="mr-1"><Plus /></el-icon>
        安装示例插件
      </el-button>
      <el-button @click="openMarket">
        <el-icon class="mr-1"><Shop /></el-icon>
        浏览市场
      </el-button>
      <el-button :loading="checkUpdateLoading" @click="checkForUpdates">
        <el-icon class="mr-1"><RefreshRight /></el-icon>
        检查更新
      </el-button>
    </div>

    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <el-table v-loading="loading" :data="plugins" stripe class="w-full">
        <el-table-column prop="name" label="插件名称" min-width="160">
          <template #default="{ row }">
            <div class="flex items-center gap-3">
              <div v-if="row.icon" class="w-10 h-10 rounded-lg bg-slate-100 flex items-center justify-center overflow-hidden shrink-0">
                <img :src="row.icon" :alt="row.name" class="w-full h-full object-cover" />
              </div>
              <div v-else class="w-10 h-10 rounded-lg bg-gradient-to-br from-blue-400 to-indigo-500 flex items-center justify-center text-white font-bold shrink-0">
                {{ row.name.charAt(0) }}
              </div>
              <div>
                <p class="font-medium text-slate-800">{{ row.name }}</p>
                <p class="text-xs text-slate-400">{{ row.plugin_id }}</p>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="100" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'warning'" size="small">
              {{ row.status === 'active' ? '运行中' : '已禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="license_key" label="License" width="120" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.license_key" size="small" type="success">已授权</el-tag>
            <el-tag v-else size="small" type="info">免费</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="installed_at" label="安装时间" width="160" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openConfig(row)">
              配置
            </el-button>
            <el-button
              v-if="row.status === 'active'"
              type="warning"
              link
              size="small"
              @click="handleDisable(row)"
            >禁用</el-button>
            <el-button
              v-if="row.status === 'disabled'"
              type="success"
              link
              size="small"
              @click="handleEnable(row)"
            >启用</el-button>
            <el-button
              v-if="updateMap[row.plugin_id]"
              type="warning"
              link
              size="small"
              @click="handleUpgrade(row)"
            >升级到 v{{ updateMap[row.plugin_id] }}</el-button>
            <el-button type="danger" link size="small" @click="handleUninstall(row)">
              卸载
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="plugins.length === 0 && !loading" class="text-center py-16 text-slate-400">
        <el-icon class="text-6xl mb-4"><Box /></el-icon>
        <p class="text-lg">暂无已安装的插件</p>
        <p class="text-sm mt-2 mb-6">点击右上角按钮安装示例插件</p>
        <el-button type="primary" @click="handleInstallDemo">
          <el-icon class="mr-1"><Plus /></el-icon>
          安装示例插件
        </el-button>
      </div>
    </div>

    <!-- 插件配置对话框 -->
    <el-dialog
      v-model="configVisible"
      :title="`${currentPlugin?.name || '插件'}配置`"
      width="560px"
    >
      <el-form v-if="currentPlugin" :model="configForm" label-width="100px">
        <template v-for="(_, key) in configForm" :key="key">
          <el-form-item :label="key">
            <el-input v-model="configForm[key]" :placeholder="`请输入${key}`" />
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="configVisible = false">取消</el-button>
        <el-button type="primary" :loading="saveConfigLoading" @click="saveConfig">
          保存配置
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="marketVisible" title="插件市场" width="750px" top="5vh">
      <div class="mb-4 flex gap-3">
        <el-input v-model="marketKeyword" placeholder="搜索插件..." clearable @keyup.enter="searchMarket" class="flex-1">
          <template #append>
            <el-button @click="searchMarket">搜索</el-button>
          </template>
        </el-input>
        <el-select v-model="marketCategory" placeholder="分类" clearable style="width: 140px" @change="searchMarket">
          <el-option v-for="cat in marketCategories" :key="cat.id" :label="cat.name" :value="cat.id" />
        </el-select>
        <el-button @click="openMarketSite" type="success">
          <el-icon class="mr-1"><Link /></el-icon> 前往市场
        </el-button>
      </div>
      <div v-loading="marketLoading">
        <div v-for="item in marketPlugins" :key="item.plugin_id" class="market-item">
          <div class="flex items-start gap-4">
            <div class="w-12 h-12 rounded-lg bg-gradient-to-br from-blue-400 to-indigo-500 flex items-center justify-center text-white font-bold text-lg shrink-0">
              {{ item.name?.charAt(0) || '?' }}
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2">
                <h4 class="font-semibold text-slate-800">{{ item.name }}</h4>
                <el-tag size="small" type="info">v{{ item.latest_version || item.version }}</el-tag>
                <el-tag v-if="item.category_name" size="small">{{ item.category_name }}</el-tag>
                <el-tag v-if="item.is_free" size="small" type="success">免费</el-tag>
              </div>
              <p class="text-sm text-slate-500 mt-1 line-clamp-2">{{ item.description }}</p>
              <p class="text-xs text-slate-400 mt-1">开发者: {{ item.author || '未知' }}</p>
            </div>
            <el-button
              :type="isInstalled(item.plugin_id) ? 'info' : 'primary'"
              size="small"
              :disabled="isInstalled(item.plugin_id)"
              @click="handleMarketInstall(item)"
            >
              {{ isInstalled(item.plugin_id) ? '已安装' : '安装' }}
            </el-button>
          </div>
        </div>
        <el-empty v-if="marketPlugins.length === 0 && !marketLoading" description="暂无可用插件" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Box, Plus, Shop, RefreshRight, Link } from '@element-plus/icons-vue'
import pluginEngineApi, { type InstalledPlugin } from '@/api/pluginEngine'

interface MarketPlugin {
  plugin_id: string
  name: string
  version: string
  latest_version?: string
  description: string
  author?: string
  category?: string
  category_name?: string
  icon?: string
  is_free?: boolean
}

interface MarketCategory {
  id: string
  name: string
  slug?: string
}

const loading = ref(false)
const plugins = ref<InstalledPlugin[]>([])

const configVisible = ref(false)
const currentPlugin = ref<InstalledPlugin | null>(null)
const configForm = reactive<Record<string, any>>({})
const saveConfigLoading = ref(false)

const marketVisible = ref(false)
const marketLoading = ref(false)
const marketKeyword = ref('')
const marketCategory = ref('')
const marketPlugins = ref<MarketPlugin[]>([])
const marketCategories = ref<MarketCategory[]>([])
const checkUpdateLoading = ref(false)
const updateMap = reactive<Record<string, string>>({})

function isInstalled(pluginId: string): boolean {
  return plugins.value.some(p => p.plugin_id === pluginId)
}

async function checkForUpdates() {
  checkUpdateLoading.value = true
  try {
    const res = await pluginEngineApi.checkUpdates()
    const updates = res.data || []
    Object.keys(updateMap).forEach(k => delete updateMap[k])
    updates.forEach((u: any) => {
      if (u.plugin_id && u.latest_version) {
        updateMap[u.plugin_id] = u.latest_version
      }
    })
    if (updates.length === 0) {
      ElMessage.success('所有插件均为最新版本')
    } else {
      ElMessage.info(`发现 ${updates.length} 个插件有更新`)
    }
  } catch (err: any) {
    ElMessage.error(err.message || '检查更新失败')
  } finally {
    checkUpdateLoading.value = false
  }
}

async function handleUpgrade(row: InstalledPlugin) {
  const newVersion = updateMap[row.plugin_id]
  if (!newVersion) return
  try {
    await ElMessageBox.confirm(
      `确定要将插件「${row.name}」从 v${row.version} 升级到 v${newVersion} 吗？`,
      '确认升级',
      { confirmButtonText: '确定升级', cancelButtonText: '取消', type: 'warning' }
    )
    await pluginEngineApi.upgradePlugin(row.plugin_id, newVersion, '')
    ElMessage.success('升级成功')
    delete updateMap[row.plugin_id]
    localStorage.setItem('menu_refresh_needed', 'true')
    fetchPlugins()
  } catch {}
}

function openMarket() {
  marketVisible.value = true
  searchMarket()
  fetchCategories()
}

async function fetchCategories() {
  try {
    const res = await pluginEngineApi.getMarketCategories()
    marketCategories.value = res.data || []
  } catch {}
}

async function openMarketSite() {
  try {
    const res = await pluginEngineApi.getSSOAuthorize()
    if (res.data?.redirect_url) {
      window.open(res.data.redirect_url, '_blank')
    }
  } catch (err: any) {
    ElMessage.error(err.message || '获取授权失败')
  }
}

async function searchMarket() {
  marketLoading.value = true
  try {
    const res = await pluginEngineApi.browseMarket(marketKeyword.value || undefined, marketCategory.value || undefined)
    const data = res.data
    if (Array.isArray(data)) {
      marketPlugins.value = data
    } else if (data?.items) {
      marketPlugins.value = data.items
    } else {
      marketPlugins.value = []
    }
  } catch (err: any) {
    marketPlugins.value = []
  } finally {
    marketLoading.value = false
  }
}

async function handleMarketInstall(item: MarketPlugin) {
  const version = item.latest_version || item.version
  try {
    const { value: licenseKey } = await ElMessageBox.prompt(
      `安装插件「${item.name}」v${version}。如有 License Key 请输入，免费插件可留空。`,
      '确认安装',
      {
        confirmButtonText: '确定安装',
        cancelButtonText: '取消',
        inputPlaceholder: 'License Key（可选）',
        inputPattern: /^$/,
        inputValidator: () => true,
      }
    )
    await pluginEngineApi.installFromMarket(item.plugin_id, version, licenseKey || '')
    ElMessage.success(`插件「${item.name}」安装成功`)
    localStorage.setItem('menu_refresh_needed', 'true')
    fetchPlugins()
  } catch {}
}

async function fetchPlugins() {
  loading.value = true
  try {
    const res = await pluginEngineApi.getInstalledPlugins()
    plugins.value = res.data || []
  } catch (err: any) {
    ElMessage.error(err.message || '获取插件列表失败')
  } finally {
    loading.value = false
  }
}

async function handleEnable(row: InstalledPlugin) {
  try {
    await pluginEngineApi.enablePlugin(row.plugin_id)
    ElMessage.success('启用成功')
    fetchPlugins()
  } catch (err: any) {
    ElMessage.error(err.message || '启用失败')
  }
}

async function handleDisable(row: InstalledPlugin) {
  try {
    await pluginEngineApi.disablePlugin(row.plugin_id)
    ElMessage.success('禁用成功')
    fetchPlugins()
  } catch (err: any) {
    ElMessage.error(err.message || '禁用失败')
  }
}

async function handleUninstall(row: InstalledPlugin) {
  try {
    await ElMessageBox.confirm(
      '确定要卸载此插件吗？卸载后将清除所有插件数据。',
      '确认卸载',
      {
        confirmButtonText: '确定卸载',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    await pluginEngineApi.uninstallPlugin(row.plugin_id)
    ElMessage.success('卸载成功')
    localStorage.setItem('menu_refresh_needed', 'true')
    fetchPlugins()
  } catch {}
}

async function handleInstallDemo() {
  try {
    await ElMessageBox.confirm(
      '确定要安装示例插件吗？安装后可在公告管理中使用。',
      '确认安装',
      { confirmButtonText: '确定安装', cancelButtonText: '取消', type: 'info' }
    )
    await pluginEngineApi.installDemo()
    ElMessage.success('示例插件安装成功')
    localStorage.setItem('menu_refresh_needed', 'true')
    fetchPlugins()
  } catch (err: any) {
    if (err?.message?.includes('已安装')) {
      ElMessage.warning('示例插件已安装，无需重复安装')
    }
  }
}

async function openConfig(row: InstalledPlugin) {
  currentPlugin.value = row
  Object.keys(configForm).forEach(key => delete configForm[key])
  if (row.config_json) {
    Object.assign(configForm, row.config_json)
  }
  configVisible.value = true
}

async function saveConfig() {
  if (!currentPlugin.value) return
  saveConfigLoading.value = true
  try {
    await pluginEngineApi.updatePluginConfig(currentPlugin.value.plugin_id, configForm)
    ElMessage.success('配置保存成功')
    configVisible.value = false
    fetchPlugins()
  } catch (err: any) {
    ElMessage.error(err.message || '保存配置失败')
  } finally {
    saveConfigLoading.value = false
  }
}

onMounted(() => {
  fetchPlugins()
})
</script>

<style scoped>
.market-item {
  padding: 16px;
  border-bottom: 1px solid #f0f0f0;
  transition: background 0.2s;
}
.market-item:hover {
  background: #f8fafc;
}
.market-item:last-child {
  border-bottom: none;
}
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
