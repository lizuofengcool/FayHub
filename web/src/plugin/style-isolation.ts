export function scopePluginCSS(css: string, pluginId: string): string {
  const prefix = `[data-plugin-container="${pluginId}"]`

  return css.replace(
    /(^|\})\s*([^{}@/]+)\s*\{/g,
    (match, closeBrace: string, selector: string) => {
      const trimmed = selector.trim()

      if (!trimmed) return match

      if (
        trimmed.startsWith('@') ||
        trimmed.startsWith(':root') ||
        trimmed.startsWith('from') ||
        trimmed.startsWith('to') ||
        trimmed.startsWith('%')
      ) {
        return match
      }

      if (trimmed === 'body' || trimmed === 'html') {
        return `${closeBrace} ${prefix} {`
      }

      if (trimmed === '*' || trimmed === '*::before' || trimmed === '*::after') {
        return `${closeBrace} ${prefix} ${trimmed} {`
      }

      const selectors = trimmed.split(',').map((s: string) => {
        const st = s.trim()
        if (st === 'body' || st === 'html') return prefix
        return `${prefix} ${st}`
      })

      return `${closeBrace} ${selectors.join(', ')} {`
    }
  )
}

export function injectPluginStyle(pluginId: string, css: string): void {
  removePluginStyle(pluginId)

  const style = document.createElement('style')
  style.id = `plugin-style-${pluginId}`
  style.setAttribute('data-plugin', pluginId)
  style.textContent = scopePluginCSS(css, pluginId)
  document.head.appendChild(style)
}

export function removePluginStyle(pluginId: string): void {
  const existing = document.getElementById(`plugin-style-${pluginId}`)
  if (existing) existing.remove()
}

export function createShadowContainer(
  hostElement: HTMLElement,
  pluginId: string
): ShadowRoot {
  const shadowRoot = hostElement.attachShadow({ mode: 'open' })

  const container = document.createElement('div')
  container.setAttribute('data-plugin-container', pluginId)
  container.id = `plugin-container-${pluginId}`
  shadowRoot.appendChild(container)

  return shadowRoot
}

export function injectStyleToShadow(
  shadowRoot: ShadowRoot,
  css: string,
  pluginId: string
): void {
  const style = document.createElement('style')
  style.setAttribute('data-plugin', pluginId)
  style.textContent = css
  shadowRoot.insertBefore(style, shadowRoot.firstChild)
}

export function injectElementPlusStylesToShadow(shadowRoot: ShadowRoot): void {
  const elStyles = document.querySelectorAll(
    'style[data-vite-dev-id*="element-plus"], link[href*="element-plus"]'
  )

  elStyles.forEach((el) => {
    if (el.tagName === 'STYLE') {
      const clone = document.createElement('style')
      clone.textContent = (el as HTMLStyleElement).textContent
      shadowRoot.insertBefore(clone, shadowRoot.firstChild)
    } else if (el.tagName === 'LINK') {
      const clone = document.createElement('link')
      clone.rel = 'stylesheet'
      clone.href = (el as HTMLLinkElement).href
      shadowRoot.insertBefore(clone, shadowRoot.firstChild)
    }
  })
}
