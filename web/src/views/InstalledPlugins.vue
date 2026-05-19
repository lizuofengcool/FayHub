<template>
  <div class="plugins-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">插件管理</h2>
          <p class="text-slate-400 text-xs mt-0.5">管理已安装的插件，开启/禁用插件</p>
        </div>
        <div class="flex gap-3">
          <el-button type="default" @click="handleInstallDemo">
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
      </div>

      <div class="p-6">
      <el-table v-loading="loading" :data="plugins" stripe class="w-full">
        <el-table-column prop="name" label="插件名称" min-width="160">
          <template #default="{ row }">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-lg bg-slate-100 flex items-center justify-center overflow-hidden shrink-0">
                <img :src="getPluginIcon(row)" :alt="getPluginDisplayName(row)" class="w-full h-full object-cover" />
              </div>
              <div>
                <p class="font-medium text-slate-800">{{ getPluginDisplayName(row) }}</p>
                <p class="text-xs text-slate-400">{{ row.plugin_id }}</p>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="100" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <n-tag :type="row.status === 'active' ? 'success' : 'warning'" size="small">
              {{ row.status === 'active' ? '运行中' : '已禁用' }}
            </n-tag>
          </template>
        </el-table-column>
        <el-table-column prop="license_key" label="License" width="120" align="center">
          <template #default="{ row }">
            <n-tag v-if="row.license_key" size="small" type="success">已授权</n-tag>
            <n-tag v-else size="small" type="default">免费</n-tag>
          </template>
        </el-table-column>
        <el-table-column prop="installed_at" label="安装时间" width="160" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="default" link size="small" @click="openConfig(row)">
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
            <el-button type="error" link size="small" @click="handleUninstall(row)">
              卸载
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="plugins.length === 0 && !loading" class="text-center py-16 text-slate-400">
        <el-icon class="text-6xl mb-4"><Box /></el-icon>
        <p class="text-lg">暂无已安装的插件</p>
        <p class="text-sm mt-2 mb-6">点击右上角按钮安装示例插件</p>
        <el-button type="default" @click="handleInstallDemo">
          <el-icon class="mr-1"><Plus /></el-icon>
          安装示例插件
        </el-button>
      </div>
      </div>
    </div>

    <!-- 插件配置对话框 -->
    <el-dialog
      v-model="configVisible"
      :title="`${currentPlugin ? getPluginDisplayName(currentPlugin) : '插件'}配置`"
      width="600px"
    >
      <el-form v-if="currentPlugin" :model="configForm" label-width="120px">
        <template v-if="Object.keys(configForm).length > 0">
          <template v-for="(value, key) in configForm" :key="key">
            <el-form-item :label="formatLabel(key)">
              <el-input
                v-if="typeof value === 'string'"
                v-model="configForm[key]"
                :placeholder="`请输入${formatLabel(key)}`"
              />
              <el-input-number
                v-else-if="typeof value === 'number'"
                v-model="configForm[key]"
                style="width: 100%"
              />
              <el-switch
                v-else-if="typeof value === 'boolean'"
                v-model="configForm[key]"
              />
              <el-input
                v-else
                :model-value="JSON.stringify(value, null, 2)"
                type="textarea"
                :rows="4"
                @update:model-value="tryParseJson(key, $event)"
              />
            </el-form-item>
          </template>
        </template>
        <n-empty v-else description="该插件暂无配置项" />
      </el-form>
      <template #footer>
        <el-button @click="configVisible = false">取消</el-button>
        <el-button type="default" :loading="saveConfigLoading" @click="saveConfig">
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
            <div class="w-12 h-12 rounded-lg bg-slate-100 flex items-center justify-center overflow-hidden shrink-0">
              <img :src="getPluginIcon(item)" :alt="getPluginDisplayName(item)" class="w-full h-full object-cover" />
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2">
                <h4 class="font-semibold text-slate-800">{{ getPluginDisplayName(item) }}</h4>
                <n-tag size="small" type="default">v{{ item.latest_version || item.version }}</n-tag>
                <n-tag v-if="item.category_name" size="small">{{ item.category_name }}</n-tag>
                <n-tag v-if="item.is_free" size="small" type="success">免费</n-tag>
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
        <n-empty v-if="marketPlugins.length === 0 && !marketLoading" description="暂无可用插件" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useMessage, useDialog } from 'naive-ui'
import { Box, Plus, Shop, RefreshRight, Link } from '@element-plus/icons-vue'
import pluginEngineApi, { type InstalledPlugin } from '@/api/pluginEngine'

const message = useMessage()
const dialog = useDialog()

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

const pluginNameMap: Record<string, string> = {
	'demo-plugin': '示例前端插件',
	'com.fayhub.announcement': '公告管理',
	'announcement': '公告管理',
	'file-manager': '文件管理',
	'user-center': '用户中心',
	'system-monitor': '系统监控',
	'api-manager': 'API管理',
	'data-encrypt-vault': '数据加密保险箱',
	'data-vault': '数据保险箱',
	'workflow-engine': '工作流引擎',
	'notify-center': '通知中心',
	'report-builder': '报表生成器',
	'form-builder': '表单设计器',
	'chat-bot': '智能客服',
	'knowledge-base': '知识库',
	'sso-connector': '单点登录连接器',
	'payment-gateway': '支付网关',
	'email-service': '邮件服务',
	'sms-service': '短信服务',
	'storage-service': '存储服务',
	'audit-log': '审计日志',
	'backup-restore': '备份恢复',
	'task-scheduler': '任务调度',
	'dashboard-widgets': '仪表盘组件',
	'theme-customizer': '主题定制器',
	'i18n-manager': '国际化管理',
	'role-manager': '角色管理',
	'permission-manager': '权限管理',
	'tenant-manager': '租户管理',
	'license-manager': '许可证管理'
}

const configLabelMap: Record<string, string> = {
  'enable': '启用',
  'title': '标题',
  'maxItems': '最大条目数',
  'apiKey': 'API密钥',
  'debugMode': '调试模式',
  'pageSize': '分页大小',
  'cacheTime': '缓存时间',
  'theme': '主题',
  'language': '语言',
  'timeout': '超时时间',
  'enabled': '已启用',
  'maxFileSize': '最大文件大小',
  'allowedTypes': '允许的文件类型',
  'uploadPath': '上传路径'
}

function getPluginDisplayName(plugin: any): string {
	const id = plugin.plugin_id || plugin.id
	return pluginNameMap[id] || plugin.name
}

function getPluginIcon(plugin: any): string {
	if (plugin.icon) {
		return plugin.icon
	}
	const id = plugin.plugin_id || plugin.id || ''
	const seed = id || 'default-plugin'
	return `https://api.dicebear.com/7.x/identicon/svg?seed=${seed}`
}

function isInstalled(pluginId: string): boolean {
  return plugins.value.some(p => p.plugin_id === pluginId)
}

async function checkForUpdates() {
  checkUpdateLoading.value = true
  try {
    const res = await pluginEngineApi.checkUpdates()
    const data = res.data
    const updates = Array.isArray(data) ? data : (data?.updates || [])
    Object.keys(updateMap).forEach(k => delete updateMap[k])
    updates.forEach((u: any) => {
      if (u.plugin_id && u.latest_version) {
        updateMap[u.plugin_id] = u.latest_version
      }
    })
    if (updates.length === 0) {
      message.success('所有插件均为最新版本')
    } else {
      message.info(`发现 ${updates.length} 个插件有更新`)
    }
  } catch (err: any) {
    message.error(err.message || '检查更新失败')
  } finally {
    checkUpdateLoading.value = false
  }
}

async function handleUpgrade(row: InstalledPlugin) {
  const newVersion = updateMap[row.plugin_id]
  if (!newVersion) return
  dialog.warning({
    title: '确认升级',
    content: `确定要将插件「${row.name}」从 v${row.version} 升级到 v${newVersion} 吗？`,
    positiveText: '确定升级',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await pluginEngineApi.upgradePlugin(row.plugin_id, newVersion, '')
        message.success('升级成功')
        delete updateMap[row.plugin_id]
        localStorage.setItem('menu_refresh_needed', 'true')
        fetchPlugins()
      } catch (e) {
        console.error('handleUpgrade failed:', e)
      }
    }
  })
}

function openMarket() {
  marketVisible.value = true
  searchMarket()
  fetchCategories()
}

async function fetchCategories() {
  try {
    const res = await pluginEngineApi.getMarketCategories()
    let rawCategories = res.data
    
    // 处理响应包装结构
    if (rawCategories?.code === 0 && rawCategories?.data) {
      rawCategories = rawCategories.data
    }
    
    if (Array.isArray(rawCategories)) {
      marketCategories.value = rawCategories.map((cat: any) => ({
        id: cat.ID || cat.id || '',
        name: cat.Name || cat.name || '',
        slug: cat.Slug || cat.slug || ''
      }))
    } else {
      marketCategories.value = []
    }
  } catch (e) { console.error('fetchCategories failed:', e); }
}

async function openMarketSite() {
  try {
    const res = await pluginEngineApi.getSSOAuthorize()
    if (res.data?.redirect_url) {
      window.open(res.data.redirect_url, '_blank')
    }
  } catch (err: any) {
    message.error(err.message || '获取授权失败')
  }
}

async function searchMarket() {
	marketLoading.value = true
	try {
		const res = await pluginEngineApi.browseMarket(marketKeyword.value || undefined, marketCategory.value || undefined)
		let responseData = res.data
		let rawPlugins: any[] = []
		
		// 处理响应包装结构
		if (responseData?.code === 0 && responseData?.data) {
			responseData = responseData.data
		}
		
		// 处理不同的响应格式
		if (Array.isArray(responseData)) {
			rawPlugins = responseData
		} else if (responseData?.list) {
			rawPlugins = responseData.list
		} else if (responseData?.items) {
			rawPlugins = responseData.items
		}
		
		// 转换字段格式，兼容后端返回的数据结构
		marketPlugins.value = rawPlugins.map((item: any) => {
			const authorName = item.developer?.name || item.developer?.teamName || 
				item.DeveloperName || item.author || '未知开发者'
			
			return {
				plugin_id: item.ID || item.plugin_id || item.id || '',
				name: item.Name || item.name || '',
				version: item.Version || item.version || '1.0.0',
				latest_version: item.latestVersion || item.LatestVersion || 
					item.latest_version || item.Version || item.version || '1.0.0',
				description: item.Description || item.description || '',
				author: authorName,
				category: item.Category || item.category || '',
				category_name: item.CategoryName || item.category_name || '',
				icon: item.CoverImage || item.coverImage || item.icon || '',
				is_free: (item.Price === 0 || item.price === 0 || item.is_free) ? true : false,
				price: item.Price || item.price || 0,
				rating: item.AverageRating || item.averageRating || 0,
				downloads: item.TotalDownloads || item.totalDownloads || 0
			}
		})
	} catch (err: any) {
		console.error('搜索Market插件失败:', err)
		marketPlugins.value = []
	} finally {
		marketLoading.value = false
	}
}

async function handleMarketInstall(item: MarketPlugin) {
  const version = item.latest_version || item.version
  const licenseKey = window.prompt(`安装插件「${item.name}」v${version}。如有 License Key 请输入，免费插件可留空。`)
  if (licenseKey === null) return
  try {
    await pluginEngineApi.installFromMarket(item.plugin_id, version, licenseKey || '')
    message.success(`插件「${item.name}」安装成功`)
    localStorage.setItem('menu_refresh_needed', 'true')
    fetchPlugins()
  } catch (e) {
    console.error('handleInstall failed:', e)
  }
}

async function fetchPlugins() {
  loading.value = true
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
  } catch (err: any) {
    message.error(err.message || '获取插件列表失败')
  } finally {
    loading.value = false
  }
}

async function handleEnable(row: InstalledPlugin) {
  try {
    await pluginEngineApi.enablePlugin(row.plugin_id)
    message.success('启用成功')
    localStorage.setItem('menu_refresh_needed', 'true')
    window.dispatchEvent(new CustomEvent('menu-refresh'))
    fetchPlugins()
  } catch (err: any) {
    message.error(err.message || '启用失败')
  }
}

async function handleDisable(row: InstalledPlugin) {
  try {
    await pluginEngineApi.disablePlugin(row.plugin_id)
    message.success('禁用成功')
    localStorage.setItem('menu_refresh_needed', 'true')
    window.dispatchEvent(new CustomEvent('menu-refresh'))
    fetchPlugins()
  } catch (err: any) {
    message.error(err.message || '禁用失败')
  }
}

async function handleUninstall(row: InstalledPlugin) {
  dialog.warning({
    title: '确认卸载',
    content: '确定要卸载此插件吗？卸载后将清除所有插件数据。',
    positiveText: '确定卸载',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await pluginEngineApi.uninstallPlugin(row.plugin_id)
        message.success('卸载成功')
        localStorage.setItem('menu_refresh_needed', 'true')
        window.dispatchEvent(new CustomEvent('menu-refresh'))
        fetchPlugins()
      } catch (e) {
        console.error('handleUninstall failed:', e)
      }
    }
  })
}

async function handleInstallDemo() {
  dialog.info({
    title: '确认安装',
    content: '确定要安装示例插件吗？安装后可在公告管理中使用。',
    positiveText: '确定安装',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await pluginEngineApi.installDemo()
        message.success('示例插件安装成功')
        localStorage.setItem('menu_refresh_needed', 'true')
        window.dispatchEvent(new CustomEvent('menu-refresh'))
        fetchPlugins()
      } catch (err: any) {
        if (err?.message?.includes('已安装')) {
          message.warning('示例插件已安装，无需重复安装')
        }
      }
    }
  })
}

const INTERNAL_MANIFEST_FIELDS = [
  'name', 'version', 'entry_point', 'description', 'min_app_version',
  'compatible_base_version', 'layout', 'render_mode', 'entry', 'style',
  'signature', 'use_shadow_dom', 'permissions', 'allowed_api_prefixes',
  'routes', 'apis', 'menus', 'page', 'config_schema',
]

function isInternalField(key: string): boolean {
  return INTERNAL_MANIFEST_FIELDS.includes(key)
}

async function openConfig(row: InstalledPlugin) {
  currentPlugin.value = row
  Object.keys(configForm).forEach(key => delete configForm[key])
  if (row.config_json) {
    try {
      const parsed = typeof row.config_json === 'string' ? JSON.parse(row.config_json) : row.config_json
      if (parsed.config_schema && typeof parsed.config_schema === 'object') {
        for (const [key, schema] of Object.entries(parsed.config_schema)) {
          if (schema && typeof schema === 'object') {
            const s = schema as Record<string, any>
            configForm[key] = s.default ?? (s.type === 'boolean' ? false : s.type === 'number' ? 0 : '')
          }
        }
      } else {
        for (const [key, value] of Object.entries(parsed)) {
          if (!isInternalField(key)) {
            configForm[key] = value
          }
        }
      }
    } catch (e) {
      console.error('解析插件配置失败:', e)
    }
  }
  configVisible.value = true
}

function formatLabel(key: string): string {
  // 先查看是否有预定义的中文标签
  if (configLabelMap[key]) {
    return configLabelMap[key]
  }
  // 将下划线、驼峰命名转换为友好的中文标签
  let label = key
    .replace(/([A-Z])/g, ' $1')
    .replace(/_/g, ' ')
    .replace(/^\w/, (c) => c.toUpperCase())
  
  // 一些常见的英文到中文的翻译
  const commonTranslations: Record<string, string> = {
    'Enable': '启用',
    'Disable': '禁用',
    'Title': '标题',
    'Name': '名称',
    'Description': '描述',
    'Url': '地址',
    'Api': '接口',
    'Key': '密钥',
    'Secret': '密钥',
    'Max': '最大',
    'Min': '最小',
    'Size': '大小',
    'Time': '时间',
    'Mode': '模式',
    'Type': '类型',
    'Path': '路径',
    'Id': '编号'
  }
  
  // 尝试翻译常见词汇
  for (const [en, zh] of Object.entries(commonTranslations)) {
    const regex = new RegExp(`\\b${en}\\b`, 'gi')
    label = label.replace(regex, zh)
  }
  
  return label
}

function tryParseJson(key: string, value: string) {
  try {
    configForm[key] = JSON.parse(value)
  } catch {
    configForm[key] = value
  }
}

async function saveConfig() {
  if (!currentPlugin.value) return
  saveConfigLoading.value = true
  try {
    await pluginEngineApi.updatePluginConfig(currentPlugin.value.plugin_id, configForm)
    message.success('配置保存成功')
    configVisible.value = false
    fetchPlugins()
  } catch (err: any) {
    message.error(err.message || '保存配置失败')
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

:deep(.el-input__wrapper) {
  height: 32px;
}

:deep(.el-select .el-input__wrapper) {
  height: 32px;
}

:deep(.el-button) {
  height: 32px;
  padding: 8px 12px;
}
</style>
