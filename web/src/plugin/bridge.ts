import request from '@/api/request'

export interface BridgeRequestOptions {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE'
  path: string
  data?: unknown
  params?: Record<string, unknown>
}

export interface BridgeUserInfo {
  id: number
  username: string
  real_name: string
  role: string
  tenant_id: number
}

export interface BridgeToastOptions {
  message: string
  type?: 'success' | 'warning' | 'info' | 'error'
  duration?: number
}

export interface BridgeConfirmOptions {
  title?: string
  message: string
  confirmText?: string
  cancelText?: string
  type?: 'success' | 'warning' | 'info' | 'error'
}

export interface BridgeNavigateOptions {
  path: string
  query?: Record<string, string>
}

type BridgeEventHandler = (data: unknown) => void

export class FayHubBridgeImpl {
  private pluginId: string
  private tenantId: number
  private allowedApiPrefixes: string[]
  private eventHandlers: Map<string, Set<BridgeEventHandler>> = new Map()

  constructor(pluginId: string, tenantId: number, allowedApiPrefixes: string[] = []) {
    this.pluginId = pluginId
    this.tenantId = tenantId
    this.allowedApiPrefixes = allowedApiPrefixes.length > 0
      ? allowedApiPrefixes
      : [`/plugin-engine/plugins/${pluginId}/`]
  }

  getPluginId(): string {
    return this.pluginId
  }

  getTenantId(): number {
    return this.tenantId
  }

  async request(options: BridgeRequestOptions): Promise<unknown> {
    this.validateApiPath(options.path)

    const config: { method: string; url: string; data?: unknown; params?: Record<string, unknown> } = {
      method: options.method.toLowerCase(),
      url: options.path,
    }
    if (options.data) config.data = options.data
    if (options.params) config.params = options.params

    const res = await request(config)
    return res.data
  }

  async getPluginConfig(): Promise<Record<string, any>> {
    const res = await request.get(`/plugin-engine/plugins/${this.pluginId}/config`)
    return res.data || {}
  }

  async setPluginConfig(config: Record<string, any>): Promise<void> {
    await request.put(`/plugin-engine/plugins/${this.pluginId}/config`, config)
  }

  async getUserInfo(): Promise<BridgeUserInfo> {
    const res = await request.get('/users/profile')
    return res.data
  }

  showToast(options: BridgeToastOptions): void {
    this.dispatchEvent('toast', options)
  }

  showConfirm(options: BridgeConfirmOptions): Promise<boolean> {
    return new Promise((resolve) => {
      this.dispatchEvent('confirm', { ...options, _resolve: resolve })
    })
  }

  navigateTo(options: BridgeNavigateOptions): void {
    this.dispatchEvent('navigate', options)
  }

  goBack(): void {
    this.dispatchEvent('navigate', { path: '__BACK__' })
  }

  on(event: string, handler: BridgeEventHandler): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, new Set())
    }
    this.eventHandlers.get(event)!.add(handler)
  }

  off(event: string, handler: BridgeEventHandler): void {
    this.eventHandlers.get(event)?.delete(handler)
  }

  dispatchEvent(event: string, data: unknown): void {
    this.eventHandlers.get(event)?.forEach(handler => {
      try {
        handler(data)
      } catch (err) {
        console.error(`[FayHubBridge] Event handler error for "${event}":`, err)
      }
    })
  }

  destroy(): void {
    this.eventHandlers.clear()
  }

  private validateApiPath(path: string): void {
    const allowed = this.allowedApiPrefixes.some(prefix => path.startsWith(prefix))
    if (!allowed) {
      throw new Error(
        `[FayHubBridge] Plugin "${this.pluginId}" is not allowed to access: ${path}. ` +
        `Allowed prefixes: ${this.allowedApiPrefixes.join(', ')}`
      )
    }
  }
}

export function createBridge(
  pluginId: string,
  tenantId: number,
  allowedApiPrefixes?: string[]
): FayHubBridgeImpl {
  return new FayHubBridgeImpl(pluginId, tenantId, allowedApiPrefixes)
}
