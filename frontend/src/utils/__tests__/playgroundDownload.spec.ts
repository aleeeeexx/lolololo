import { beforeEach, describe, expect, it, vi } from 'vitest'
import { blobFromImageSource, downloadPlaygroundImage } from '@/utils/playgroundDownload'

function pngBytes(): string {
  const bytes = new Uint8Array([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a])
  let binary = ''
  bytes.forEach((value) => {
    binary += String.fromCharCode(value)
  })
  return btoa(binary)
}

describe('playground image download', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
  })

  it('builds a file blob from b64_json', () => {
    const blob = blobFromImageSource({ b64_json: pngBytes() })
    expect(blob).toBeInstanceOf(Blob)
    expect(blob?.type).toBe('image/png')
  })

  it('builds a file blob from a data URL', () => {
    const blob = blobFromImageSource({ url: `data:image/png;base64,${pngBytes()}` })
    expect(blob).toBeInstanceOf(Blob)
    expect(blob?.type).toBe('image/png')
  })

  it('falls back to the authenticated content endpoint when a remote URL cannot be fetched', async () => {
    vi.stubGlobal('URL', {
      createObjectURL: vi.fn(() => 'blob:test'),
      revokeObjectURL: vi.fn(),
    })
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('cors')))
    const click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})
    const fallback = vi.fn().mockResolvedValue(new Blob([new Uint8Array([1, 2, 3])], { type: 'image/png' }))
    await downloadPlaygroundImage({
      source: { url: 'https://cdn.example/test.png' },
      filename: 'result.png',
      fallback,
    })
    expect(fallback).toHaveBeenCalledOnce()
    expect(click).toHaveBeenCalledOnce()
    click.mockRestore()
    vi.unstubAllGlobals()
  })
})
