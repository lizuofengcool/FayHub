import { defineAsyncComponent, type Component } from 'vue'
import { PluginSandbox } from './sandbox'
import { createBridge, type FayHubBridgeImpl } from './bridge'
import { verifyPluginSignature, getMarketPublicKey } from './signature'
import { injectPluginStyle, removePluginStyle } from './style-isolation'

export interface PluginManifest {
  id: string
  name: string
  version: string
  entry: string
  style?: string
  renderMode: 'custom' | 'schema'
  permissions?: string[]
  allowedApiPrefixes?: string[]
  compatibleBaseVersion?: string
  signature?: string
}

export interface LoadedPlugin {
  pluginId: string
  manifest: PluginManifest
  sandbox: PluginSandbox
  bridge: FayHubBridgeImpl
  component: Component | null
  status: 'loading' | 'loaded' | 'error'
  error?: string
}

const pluginCache = new Map<string, LoadedPlugin>()

const BASE_VERSION = '1.0.0'

export async function loadPlugin(
  manifest: PluginManifest,
  tenantId: number
): Promise<LoadedPlugin> {
  const cached = pluginCache.get(manifest.id)
  if (cached && cached.status === 'loaded') {
    return cached
  }

  if (manifest.compatibleBaseVersion) {
    const compatible = checkVersionCompatibility(
      BASE_VERSION,
      manifest.compatibleBaseVersion
    )
    if (!compatible) {
      return {
        pluginId: manifest.id,
        manifest,
        sandbox: null as any,
        bridge: null as any,
        component: null,
        status: 'error',
        error: `Plugin requires base version ${manifest.compatibleBaseVersion}, current is ${BASE_VERSION}`,
      }
    }
  }

  const bridge = createBridge(
    manifest.id,
    tenantId,
    manifest.allowedApiPrefixes
  )

  const sandbox = new PluginSandbox({
    pluginId: manifest.id,
    bridge: bridge as any,
  })

  const loaded: LoadedPlugin = {
    pluginId: manifest.id,
    manifest,
    sandbox,
    bridge,
    component: null,
    status: 'loading',
  }

  pluginCache.set(manifest.id, loaded)

  try {
    const jsCode = await fetchPluginAsset(manifest.id, manifest.entry, manifest.version)

    if (manifest.signature) {
      const publicKey = await getMarketPublicKey()
      if (!publicKey) {
        throw new Error(
          'Plugin signature verification failed: unable to fetch market public key'
        )
      }
      const encoder = new TextEncoder()
      const payload = encoder.encode(jsCode)
      const sigResult = await verifyPluginSignature(
        payload.buffer as ArrayBuffer,
        manifest.signature,
        publicKey
      )
      if (!sigResult.valid) {
        throw new Error(
          `Plugin signature verification failed: ${sigResult.error || 'Invalid signature'}`
        )
      }
    }

    const component = sandbox.executeScript(jsCode)

    if (!component) {
      throw new Error('Plugin did not export a component')
    }

    if (
      typeof component !== 'function' &&
      typeof component !== 'object'
    ) {
      throw new Error('Plugin export is not a valid Vue component')
    }

    if (manifest.style) {
      const cssCode = await fetchPluginAsset(manifest.id, manifest.style, manifest.version)
      injectPluginStyle(manifest.id, cssCode)
    }

    loaded.component = wrapWithAsyncComponent(component, manifest.id)
    loaded.status = 'loaded'
  } catch (err: any) {
    loaded.status = 'error'
    loaded.error = err.message || 'Unknown error'
    console.error(`[PluginLoader] Failed to load "${manifest.id}":`, err)
  }

  return loaded
}

export function unloadPlugin(pluginId: string): void {
  const loaded = pluginCache.get(pluginId)
  if (!loaded) return

  loaded.sandbox.destroy()
  loaded.bridge.destroy()
  removePluginStyle(pluginId)
  pluginCache.delete(pluginId)
}

export function getLoadedPlugin(pluginId: string): LoadedPlugin | undefined {
  return pluginCache.get(pluginId)
}

export function getAllLoadedPlugins(): LoadedPlugin[] {
  return Array.from(pluginCache.values())
}

async function fetchPluginAsset(
  pluginId: string,
  assetPath: string,
  version?: string
): Promise<string> {
  const cacheBuster = version ? `?v=${encodeURIComponent(version)}` : `?t=${Date.now()}`
  const url = `/plugin-assets/${pluginId}/${assetPath}${cacheBuster}`
  const response = await fetch(url)
  if (!response.ok) {
    throw new Error(
      `Failed to fetch plugin asset: ${url} (${response.status})`
    )
  }
  return response.text()
}

function wrapWithAsyncComponent(
  component: Component,
  pluginId: string
): Component {
  return defineAsyncComponent({
    loader: () => Promise.resolve(component),
    loadingComponent: {
      name: `PluginLoading_${pluginId}`,
      template: `<div class="flex items-center justify-center py-20"><p class="text-slate-400 animate-pulse">Loading plugin...</p></div>`,
    },
    errorComponent: {
      name: `PluginError_${pluginId}`,
      props: ['error'],
      template: `<div class="flex flex-col items-center justify-center py-20"><p class="text-red-500 text-lg mb-2">Plugin load failed</p><p class="text-slate-400 text-sm">{{ error }}</p></div>`,
    },
    delay: 0,
    timeout: 10000,
  })
}

function checkVersionCompatibility(
  currentVersion: string,
  requiredRange: string
): boolean {
  const current = parseVersion(currentVersion)
  if (!current) return true

  const rangeMatch = requiredRange.match(
    /^>=?([\d.]+)\s*<?([\d.]+)?$/
  )
  if (!rangeMatch) return true

  const min = parseVersion(rangeMatch[1])
  if (min && compareVersions(current, min) < 0) return false

  if (rangeMatch[2]) {
    const max = parseVersion(rangeMatch[2])
    if (max && compareVersions(current, max) >= 0) return false
  }

  return true
}

function parseVersion(v: string): [number, number, number] | null {
  const parts = v.split('.').map(Number)
  if (parts.length < 3 || parts.some(isNaN)) return null
  return parts as [number, number, number]
}

function compareVersions(
  a: [number, number, number],
  b: [number, number, number]
): number {
  for (let i = 0; i < 3; i++) {
    if (a[i] !== b[i]) return a[i] - b[i]
  }
  return 0
}
