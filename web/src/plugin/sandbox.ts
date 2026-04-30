const RAW_WINDOW = window

const ALLOWED_GLOBALS: Set<string> = new Set([
  'undefined', 'NaN', 'Infinity',
  'parseInt', 'parseFloat', 'isNaN', 'isFinite',
  'decodeURI', 'decodeURIComponent', 'encodeURI', 'encodeURIComponent',
  'Object', 'Array', 'String', 'Boolean', 'Number', 'Symbol', 'BigInt',
  'Map', 'Set', 'WeakMap', 'WeakSet', 'WeakRef',
  'Date', 'RegExp', 'Error', 'TypeError', 'RangeError', 'SyntaxError',
  'JSON', 'Math', 'Promise', 'Proxy', 'Reflect',
  'console', 'setTimeout', 'setInterval', 'clearTimeout', 'clearInterval',
  'requestAnimationFrame', 'cancelAnimationFrame',
  'Intl', 'TextEncoder', 'TextDecoder',
  'URL', 'URLSearchParams',
  'Event', 'CustomEvent', 'AbortController', 'AbortSignal',
  'Blob', 'File', 'FileReader', 'FormData',
  'Headers', 'Request', 'Response',
  'crypto',
])

const BLOCKED_PROPERTIES: Set<string> = new Set([
  'localStorage', 'sessionStorage', 'indexedDB',
  'cookieStore', 'caches',
  'open', 'close', 'focus', 'blur',
  'postMessage', 'opener', 'parent', 'frames', 'top',
  'document',
])

export interface SandboxOptions {
  pluginId: string
  bridge: Record<string, any>
}

export class PluginSandbox {
  private pluginId: string
  private proxyWindow: Window & typeof globalThis
  private bridge: Record<string, any>
  private fakeDocument: Document
  private destroyed: boolean = false

  constructor(options: SandboxOptions) {
    this.pluginId = options.pluginId
    this.bridge = options.bridge
    this.fakeDocument = this.createFakeDocument()
    this.proxyWindow = this.createProxyWindow()
  }

  getProxyWindow(): Window & typeof globalThis {
    return this.proxyWindow
  }

  getPluginId(): string {
    return this.pluginId
  }

  executeScript(code: string): any {
    if (this.destroyed) {
      throw new Error(`[FayHub Sandbox] Plugin "${this.pluginId}" sandbox has been destroyed`)
    }

    const moduleExports: Record<string, any> = {}
    const moduleObj = { exports: moduleExports }

    const wrappedCode = `
      (function(module, exports, require, __filename, __dirname) {
        ${code}
      })
    `

    const fn = new Function('return ' + wrappedCode)()

    const fakeRequire = (dep: string) => {
      if (dep === 'vue') {
        if ((RAW_WINDOW as any).Vue) return (RAW_WINDOW as any).Vue
        const vueModule = (RAW_WINDOW as any).__VUE_DEVTOOLS_GLOBAL_HOOK__
        if (vueModule) return vueModule
        throw new Error(
          `[FayHub Sandbox] Plugin "${this.pluginId}": Vue is not available on global scope. ` +
          `Ensure the host app registers Vue globally before loading plugins.`
        )
      }
      if (dep === 'element-plus') {
        if ((RAW_WINDOW as any).ElementPlus) return (RAW_WINDOW as any).ElementPlus
        throw new Error(
          `[FayHub Sandbox] Plugin "${this.pluginId}": ElementPlus is not available on global scope.`
        )
      }
      if (dep === '@element-plus/icons-vue') {
        if ((RAW_WINDOW as any).ElementPlusIconsVue) return (RAW_WINDOW as any).ElementPlusIconsVue
        throw new Error(
          `[FayHub Sandbox] Plugin "${this.pluginId}": ElementPlusIconsVue is not available on global scope.`
        )
      }
      if (dep === '@fayhub/bridge') {
        return this.bridge
      }
      throw new Error(
        `[FayHub Sandbox] Plugin "${this.pluginId}" requires unknown dependency: ${dep}`
      )
    }

    fn.call(
      this.proxyWindow,
      moduleObj,
      moduleObj.exports,
      fakeRequire,
      `${this.pluginId}/index.js`,
      this.pluginId
    )

    return moduleObj.exports.default || moduleObj.exports
  }

  destroy(): void {
    this.destroyed = true
    this.bridge = {}
  }

  isDestroyed(): boolean {
    return this.destroyed
  }

  private createProxyWindow(): Window & typeof globalThis {
    const sandbox = this
    const fakeWindow: Record<string, any> = {}

    const proxy = new Proxy(fakeWindow, {
      get(target: Record<string, any>, prop: string | symbol): any {
        const key = String(prop)

        if (sandbox.destroyed) return undefined

        if (key === 'window' || key === 'self' || key === 'globalThis') {
          return proxy
        }

        if (key === 'document') {
          return sandbox.fakeDocument
        }

        if (key === '__FAYHUB_BRIDGE__' || key === 'FayHubBridge') {
          return sandbox.bridge
        }

        if (target[key] !== undefined) {
          return target[key]
        }

        if (ALLOWED_GLOBALS.has(key)) {
          return (RAW_WINDOW as any)[key]
        }

        if (BLOCKED_PROPERTIES.has(key)) {
          console.warn(
            `[FayHub Sandbox] Plugin "${sandbox.pluginId}" blocked access: ${key}`
          )
          return undefined
        }

        return (RAW_WINDOW as any)[key]
      },

      set(target: Record<string, any>, prop: string | symbol, value: any): boolean {
        const key = String(prop)

        if (BLOCKED_PROPERTIES.has(key)) {
          console.warn(
            `[FayHub Sandbox] Plugin "${sandbox.pluginId}" blocked write: ${key}`
          )
          return true
        }

        target[key] = value
        return true
      },

      has(target: Record<string, any>, prop: string | symbol): boolean {
        const key = String(prop)
        if (BLOCKED_PROPERTIES.has(key)) return false
        return key in target || key in RAW_WINDOW
      },

      deleteProperty(target: Record<string, any>, prop: string | symbol): boolean {
        const key = String(prop)
        if (BLOCKED_PROPERTIES.has(key)) return true
        delete target[key]
        return true
      }
    })

    return proxy as Window & typeof globalThis
  }

  private createFakeDocument(): Document {
    const sandbox = this
    const realDoc = RAW_WINDOW.document

    const handler: ProxyHandler<Document> = {
      get(_target, prop) {
        const key = String(prop)

        if (key === 'title') return `Plugin: ${sandbox.pluginId}`
        if (key === 'cookie') return ''

        if (key === 'createElement') {
          return function(tag: string) {
            return realDoc.createElement(tag)
          }
        }

        if (key === 'createTextNode') {
          return function(text: string) {
            return realDoc.createTextNode(text)
          }
        }

        if (key === 'getElementById') {
          return function(id: string) {
            return realDoc.getElementById(id)
          }
        }

        if (key === 'querySelector' || key === 'querySelectorAll') {
          return function(selector: string) {
            const container = realDoc.getElementById(
              `plugin-container-${sandbox.pluginId}`
            )
            if (!container) {
              return key === 'querySelectorAll' ? [] : null
            }
            return container[key](selector)
          }
        }

        if (key === 'addEventListener' || key === 'removeEventListener') {
          return function(type: string, listener: any, options?: any) {
            const container = realDoc.getElementById(
              `plugin-container-${sandbox.pluginId}`
            )
            if (container) {
              container[key](type, listener, options)
            }
          }
        }

        const blocked = ['write', 'writeln', 'domain']
        if (blocked.includes(key)) {
          console.warn(
            `[FayHub Sandbox] Plugin "${sandbox.pluginId}" blocked document.${key}`
          )
          return () => {}
        }

        if (typeof (realDoc as any)[key] === 'function') {
          return (realDoc as any)[key].bind(realDoc)
        }

        return (realDoc as any)[key]
      },

      set(_target, prop, value) {
        const key = String(prop)
        if (key === 'title' || key === 'cookie') return true
        ;(realDoc as any)[key] = value
        return true
      }
    }

    return new Proxy({} as Document, handler)
  }
}
