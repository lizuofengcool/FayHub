import request from './request'

export interface PluginRuntimeStats {
  plugin_id: string
  tenant_id: number
  call_count: number
  error_count: number
  total_duration_ms: number
  max_duration_ms: number
  avg_duration_ms: number
  last_call_at: string
  last_error_at: string | null
  last_error_msg: string
  memory_usage_kb: number
  cpu_percent: number
  status: string
}

export interface PluginDBStats {
  id: number
  tenant_id: number
  plugin_id: string
  call_count: number
  error_count: number
  total_duration_ms: number
  max_duration_ms: number
  last_call_at: string
  last_error_at: string | null
  last_error_msg: string
  memory_usage_kb: number
  status: string
}

export interface PluginAlert {
  id: number
  plugin_id: string
  event_type: string
  event_data: string
  created_at: string
}

export interface StatsSummary {
  total_calls: number
  total_errors: number
  total_plugins: number
  error_rate: number
  updated_at: string
}

const pluginMonitorApi = {
  getRuntimeStats(pluginId?: string) {
    const params: Record<string, string> = {}
    if (pluginId) params.plugin_id = pluginId
    return request.get('/plugin-monitor/runtime', { params })
  },

  getDBStats(pluginId?: string) {
    const params: Record<string, string> = {}
    if (pluginId) params.plugin_id = pluginId
    return request.get('/plugin-monitor/db', { params })
  },

  getAlerts(pluginId?: string) {
    const params: Record<string, string> = {}
    if (pluginId) params.plugin_id = pluginId
    return request.get('/plugin-monitor/alerts', { params })
  },

  resetStats(pluginId: string) {
    return request.post('/plugin-monitor/reset', null, {
      params: { plugin_id: pluginId }
    })
  }
}

export default pluginMonitorApi
