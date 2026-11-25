export interface PlaygroundImageSource {
  url?: string
  b64_json?: string
}

function extensionForMime(mime: string): string {
  const normalized = mime.toLowerCase()
  if (normalized.includes('jpeg') || normalized.includes('jpg')) return 'jpg'
  if (normalized.includes('webp')) return 'webp'
  if (normalized.includes('gif')) return 'gif'
  return 'png'
}

function decodeBase64(value: string): Uint8Array {
  const trimmed = value.trim()
  const payload = trimmed.includes('base64,') ? trimmed.slice(trimmed.indexOf('base64,') + 7) : trimmed
  const binary = atob(payload)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  return bytes
}

export function blobFromImageSource(source: PlaygroundImageSource): Blob | null {
  if (source.b64_json && source.b64_json.trim()) {
    const bytes = decodeBase64(source.b64_json)
    const mime =
      bytes.length >= 3 && bytes[0] === 0xff && bytes[1] === 0xd8 && bytes[2] === 0xff
        ? 'image/jpeg'
        : bytes.length >= 12 &&
            String.fromCharCode(...bytes.slice(0, 4)) === 'RIFF' &&
            String.fromCharCode(...bytes.slice(8, 12)) === 'WEBP'
          ? 'image/webp'
          : 'image/png'
    return new Blob([bytes], { type: mime })
  }
  const url = source.url?.trim() || ''
  if (url.startsWith('data:image/')) {
    const bytes = decodeBase64(url)
    const mime = url.slice(5, url.indexOf(';')) || 'image/png'
    return new Blob([bytes], { type: mime })
  }
  return null
}

export function downloadBlob(blob: Blob, filename: string): void {
  const objectUrl = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = objectUrl
  anchor.download = filename
  anchor.rel = 'noopener'
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
  window.setTimeout(() => URL.revokeObjectURL(objectUrl), 1000)
}

export async function downloadPlaygroundImage(options: {
  source: PlaygroundImageSource
  filename: string
  fallback?: () => Promise<Blob>
}): Promise<void> {
  const local = blobFromImageSource(options.source)
  if (local) {
    const ext = extensionForMime(local.type)
    const base = options.filename.replace(/\.(png|jpe?g|webp|gif)$/i, '')
    downloadBlob(local, `${base}.${ext}`)
    return
  }
  const url = options.source.url?.trim() || ''
  if (url && /^https?:/i.test(url)) {
    try {
      const response = await fetch(url)
      if (response.ok) {
        const blob = await response.blob()
        const ext = extensionForMime(blob.type || 'image/png')
        const base = options.filename.replace(/\.(png|jpe?g|webp|gif)$/i, '')
        downloadBlob(blob, `${base}.${ext}`)
        return
      }
    } catch {
      // Fall through to the authenticated content endpoint.
    }
  }
  if (!options.fallback) {
    throw new Error('image is not downloadable')
  }
  const blob = await options.fallback()
  const ext = extensionForMime(blob.type || 'image/png')
  const base = options.filename.replace(/\.(png|jpe?g|webp|gif)$/i, '')
  downloadBlob(blob, `${base}.${ext}`)
}
