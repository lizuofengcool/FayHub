import request from './request'

export function getEngineStatus() {
  return request.get('/engine/status')
}

export function getLoadedPlugins() {
  return request.get('/engine/plugins')
}

export function getPluginRoutes(pluginId: string) {
  return request.get(`/engine/plugins/${pluginId}/routes`)
}

export function healthCheckPlugin(pluginId: string) {
  return request.get(`/engine/plugins/${pluginId}/health`)
}
