import request from './request'

export interface InstalledPlugin {
  id: number
  plugin_id: string
  name: string
  version: string
  icon: string
  description: string
  config_json: Record<string, any>
  license_key: string
  status: 'active' | 'disabled'
  installed_at: string
  updated_at: string
}

export interface InstallPluginRequest {
  plugin_id: string
  version: string
  license_key: string
}

const pluginEngineApi = {
  // 获取已安装的插件列表
  getInstalledPlugins() {
    return request.get('/plugin-engine/plugins')
  },

  // 获取单个插件详情
  getPlugin(pluginId: string) {
    return request.get(`/plugin-engine/plugins/${pluginId}`)
  },

  // 插件安装回调（从市场跳回来时调用）
  installCallback(data: InstallPluginRequest) {
    return request.post('/plugin-engine/install-callback', data)
  },

  // 卸载插件
  uninstallPlugin(pluginId: string) {
    return request.delete(`/plugin-engine/plugins/${pluginId}`)
  },

  // 启用插件
  enablePlugin(pluginId: string) {
    return request.put(`/plugin-engine/plugins/${pluginId}/enable`)
  },

  // 禁用插件
  disablePlugin(pluginId: string) {
    return request.put(`/plugin-engine/plugins/${pluginId}/disable`)
  },

  // 升级插件
  upgradePlugin(pluginId: string, newVersion: string, newLicenseKey: string) {
    return request.put(`/plugin-engine/plugins/${pluginId}/upgrade`, {
      new_version: newVersion,
      new_license_key: newLicenseKey
    })
  },

  // 获取插件配置
  getPluginConfig(pluginId: string) {
    return request.get(`/plugin-engine/plugins/${pluginId}/config`)
  },

  // 更新插件配置
  updatePluginConfig(pluginId: string, config: Record<string, any>) {
    return request.put(`/plugin-engine/plugins/${pluginId}/config`, config)
  },

  // 安装示例插件
  installDemo() {
    return request.post('/plugin-engine/demo/install')
  },

  browseMarket(keyword?: string, category?: string) {
    const params: Record<string, string> = {}
    if (keyword) params.keyword = keyword
    if (category) params.category = category
    return request.get('/plugin-engine/market/plugins', { params })
  },

  installFromMarket(pluginId: string, version: string, licenseKey?: string) {
    return request.post('/plugin-engine/install-callback', {
      plugin_id: pluginId,
      version,
      license_key: licenseKey || '',
    })
  },

  checkUpdates() {
    return request.get('/plugin-engine/plugins/check-updates')
  },
}

export default pluginEngineApi
