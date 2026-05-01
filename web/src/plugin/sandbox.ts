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
  private trackedTimers: Set<ReturnType<typeof setTimeout> | ReturnType<typeof setInterval>> = new Set()
  private trackedRAFs: Set<number> = new Set()
  private trackedEventListeners: Array<{ target: EventTarget; type: string; listener: EventListenerOrEventListenerObject }> = []
  private trackedObservers: Set<MutationObserver | ResizeObserver> = new Set()

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
        "use strict";
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
    this.clearAllTimers()
    this.removeAllEventListeners()
    this.disconnectAllObservers()
    this.bridge = {}
  }

  isDestroyed(): boolean {
    return this.destroyed
  }

  clearAllTimers(): void {
    this.trackedTimers.forEach(id => {
      clearTimeout(id as ReturnType<typeof setTimeout>)
      clearInterval(id as ReturnType<typeof setInterval>)
    })
    this.trackedTimers.clear()

    this.trackedRAFs.forEach(id => {
      cancelAnimationFrame(id)
    })
    this.trackedRAFs.clear()
  }

  removeAllEventListeners(): void {
    this.trackedEventListeners.forEach(({ target, type, listener }) => {
      target.removeEventListener(type, listener)
    })
    this.trackedEventListeners = []
  }

  disconnectAllObservers(): void {
    this.trackedObservers.forEach(obs => {
      obs.disconnect()
    })
    this.trackedObservers.clear()
  }

  trackTimer(id: ReturnType<typeof setTimeout> | ReturnType<typeof setInterval>): void {
    this.trackedTimers.add(id)
  }

  untrackTimer(id: ReturnType<typeof setTimeout> | ReturnType<typeof setInterval>): void {
    this.trackedTimers.delete(id)
  }

  trackRAF(id: number): void {
    this.trackedRAFs.add(id)
  }

  untrackRAF(id: number): void {
    this.trackedRAFs.delete(id)
  }

  trackEventListener(target: EventTarget, type: string, listener: EventListenerOrEventListenerObject): void {
    this.trackedEventListeners.push({ target, type, listener })
  }

  trackObserver(obs: MutationObserver | ResizeObserver): void {
    this.trackedObservers.add(obs)
  }

  private createProxyWindow(): Window & typeof globalThis {
    const sandbox = this
    const fakeWindow: Record<string, any> = {}

    const originalSetTimeout = RAW_WINDOW.setTimeout.bind(RAW_WINDOW)
    const originalSetInterval = RAW_WINDOW.setInterval.bind(RAW_WINDOW)
    const originalClearTimeout = RAW_WINDOW.clearTimeout.bind(RAW_WINDOW)
    const originalClearInterval = RAW_WINDOW.clearInterval.bind(RAW_WINDOW)
    const originalRAF = RAW_WINDOW.requestAnimationFrame.bind(RAW_WINDOW)
    const originalCancelRAF = RAW_WINDOW.cancelAnimationFrame.bind(RAW_WINDOW)

    fakeWindow.setTimeout = function(fn: Function, delay?: number, ...args: any[]) {
      const id = originalSetTimeout(fn, delay, ...args)
      sandbox.trackTimer(id as ReturnType<typeof setTimeout>)
      return id
    }
    fakeWindow.setInterval = function(fn: Function, delay?: number, ...args: any[]) {
      const id = originalSetInterval(fn, delay, ...args)
      sandbox.trackTimer(id as ReturnType<typeof setInterval>)
      return id
    }
    fakeWindow.clearTimeout = function(id?: ReturnType<typeof setTimeout>) {
      if (id !== undefined) {
        sandbox.untrackTimer(id)
        originalClearTimeout(id)
      }
    }
    fakeWindow.clearInterval = function(id?: ReturnType<typeof setInterval>) {
      if (id !== undefined) {
        sandbox.untrackTimer(id)
        originalClearInterval(id)
      }
    }
    fakeWindow.requestAnimationFrame = function(callback: FrameRequestCallback) {
      const id = originalRAF(callback)
      sandbox.trackRAF(id)
      return id
    }
    fakeWindow.cancelAnimationFrame = function(id?: number) {
      if (id !== undefined) {
        sandbox.untrackRAF(id)
        originalCancelRAF(id)
      }
    }

    const SafeFunction = function(...args: string[]) {
      if (args.length > 0 && typeof args[args.length - 1] === 'string') {
        const body = args[args.length - 1] as string
        if (body.includes('return this') || body.includes('return(globalThis)') || body.includes('return globalThis')) {
          console.warn(`[FayHub Sandbox] Plugin "${sandbox.pluginId}" attempted sandbox escape via Function constructor`)
          return function() { return undefined }
        }
      }
      return RAW_WINDOW.Function(...args)
    }
    SafeFunction.prototype = RAW_WINDOW.Function.prototype
    ;(SafeFunction as any).constructor = SafeFunction
    fakeWindow.Function = SafeFunction

    const OriginalError = RAW_WINDOW.Error
    const SafeError = function(this: any, ...args: any[]) {
      const err = new OriginalError(...args)
      if (err.stack) {
        err.stack = err.stack
          .split('\n')
          .filter((line: string) => !line.includes('fayhub') && !line.includes('plugin-assets'))
          .join('\n')
      }
      return err
    } as any
    SafeError.prototype = OriginalError.prototype
    SafeError.captureStackTrace = OriginalError.captureStackTrace
    SafeError.stackTraceLimit = OriginalError.stackTraceLimit
    fakeWindow.Error = SafeError

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
          const raw = (RAW_WINDOW as any)[key]
          if (key === 'Function') return SafeFunction
          if (key === 'Error') return SafeError
          if (key === 'setTimeout') return fakeWindow.setTimeout
          if (key === 'setInterval') return fakeWindow.setInterval
          if (key === 'clearTimeout') return fakeWindow.clearTimeout
          if (key === 'clearInterval') return fakeWindow.clearInterval
          if (key === 'requestAnimationFrame') return fakeWindow.requestAnimationFrame
          if (key === 'cancelAnimationFrame') return fakeWindow.cancelAnimationFrame
          return raw
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

        if (key === 'addEventListener') {
          return function(type: string, listener: EventListenerOrEventListenerObject, options?: boolean | AddEventListenerOptions) {
            const container = realDoc.getElementById(
              `plugin-container-${sandbox.pluginId}`
            )
            if (container) {
              container.addEventListener(type, listener, options)
              sandbox.trackEventListener(container, type, listener)
            }
          }
        }

        if (key === 'removeEventListener') {
          return function(type: string, listener: EventListenerOrEventListenerObject, options?: boolean | EventListenerOptions) {
            const container = realDoc.getElementById(
              `plugin-container-${sandbox.pluginId}`
            )
            if (container) {
              container.removeEventListener(type, listener, options)
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
