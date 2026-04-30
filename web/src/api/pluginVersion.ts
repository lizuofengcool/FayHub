import request, { type PageResult } from './request'

export interface PluginVersion {
  id: number
  plugin_id: string
  version: string
  changelog: string
  download_url: string
  signature: string
  min_engine_version: string
  is_latest: boolean
  created_at: string
}

export interface PluginVersionHistory {
  id: number
  plugin_id: string
  from_version: string
  to_version: string
  action: string
  operator: string
  created_at: string
}

export interface PluginDependency {
  id: number
  plugin_id: string
  dependency_plugin_id: string
  dependency_version: string
  is_required: boolean
}

const pluginVersionApi = {
  listVersions(pluginId: string) {
    return request.get<any, { code: number; data: PluginVersion[] }>(`/plugin-engine/plugins/${pluginId}/versions`)
  },

  getVersion(pluginId: string, version: string) {
    return request.get<any, { code: number; data: PluginVersion }>(`/plugin-engine/plugins/${pluginId}/versions/${version}`)
  },

  listVersionHistory(pluginId: string, params?: { page?: number; page_size?: number }) {
    return request.get<any, { code: number; data: PageResult<PluginVersionHistory> }>(`/plugin-engine/plugins/${pluginId}/history`, { params })
  },

  listDependencies(pluginId: string) {
    return request.get<any, { code: number; data: PluginDependency[] }>(`/plugin-engine/plugins/${pluginId}/dependencies`)
  },

  checkUpdates(pluginId: string) {
    return request.get<any, { code: number; data: { has_update: boolean; latest_version: string; changelog: string } }>(`/plugin-engine/plugins/${pluginId}/check-update`)
  }
}

export default pluginVersionApi
