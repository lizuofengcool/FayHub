export interface PluginSignatureResult {
  valid: boolean
  error?: string
}

export async function verifyPluginSignature(
  payload: ArrayBuffer,
  signatureB64: string,
  publicKeyPem: string
): Promise<PluginSignatureResult> {
  try {
    const publicKey = await importPublicKey(publicKeyPem)

    const sigBuffer = base64ToArrayBuffer(signatureB64)

    const isValid = await crypto.subtle.verify(
      { name: 'RSASSA-PKCS1-v1_5', hash: 'SHA-256' },
      publicKey,
      sigBuffer,
      payload
    )

    return { valid: isValid }
  } catch (err: unknown) {
    return { valid: false, error: err instanceof Error ? err.message : 'Signature verification failed' }
  }
}

export async function computeSHA256(data: ArrayBuffer): Promise<string> {
  const hashBuffer = await crypto.subtle.digest('SHA-256', data)
  return arrayBufferToHex(hashBuffer)
}

async function importPublicKey(pem: string): Promise<CryptoKey> {
  const pemBody = pem
    .replace(/-----BEGIN PUBLIC KEY-----/, '')
    .replace(/-----END PUBLIC KEY-----/, '')
    .replace(/\s/g, '')

  const binaryStr = atob(pemBody)
  const bytes = new Uint8Array(binaryStr.length)
  for (let i = 0; i < binaryStr.length; i++) {
    bytes[i] = binaryStr.charCodeAt(i)
  }

  return crypto.subtle.importKey(
    'spki',
    bytes.buffer,
    { name: 'RSASSA-PKCS1-v1_5', hash: 'SHA-256' },
    false,
    ['verify']
  )
}

function base64ToArrayBuffer(b64: string): ArrayBuffer {
  const binaryStr = atob(b64)
  const bytes = new Uint8Array(binaryStr.length)
  for (let i = 0; i < binaryStr.length; i++) {
    bytes[i] = binaryStr.charCodeAt(i)
  }
  return bytes.buffer
}

function arrayBufferToHex(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer)
  return Array.from(bytes)
    .map(b => b.toString(16).padStart(2, '0'))
    .join('')
}

let cachedPublicKey: string | null = null

export async function getMarketPublicKey(): Promise<string> {
  if (cachedPublicKey) return cachedPublicKey

  try {
    const res = await fetch('/api/plugin-engine/market/public-key')
    if (res.ok) {
      const data = await res.json()
      cachedPublicKey = data.data?.public_key || data.public_key || ''
      return cachedPublicKey
    }
  } catch (e) { console.error('fetchPublicKey failed:', e); }

  return ''
}
