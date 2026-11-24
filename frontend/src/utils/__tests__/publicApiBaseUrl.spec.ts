import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { resolvePublicApiBaseUrl } from '../publicApiBaseUrl'

describe('resolvePublicApiBaseUrl', () => {
  const originalLocation = window.location

  beforeEach(() => {
    Object.defineProperty(window, 'location', {
      value: { origin: 'https://app.example.com' },
      writable: true,
      configurable: true,
    })
  })

  afterEach(() => {
    Object.defineProperty(window, 'location', {
      value: originalLocation,
      writable: true,
      configurable: true,
    })
  })

  it('uses the configured API base URL when present', () => {
    expect(resolvePublicApiBaseUrl('https://api.example.com')).toBe('https://api.example.com')
  })

  it('trims whitespace and trailing slashes from the configured URL', () => {
    expect(resolvePublicApiBaseUrl('  https://api.example.com/v1/  ')).toBe('https://api.example.com/v1')
  })

  it('preserves a configured trailing /v1 instead of stripping or inventing paths', () => {
    expect(resolvePublicApiBaseUrl('https://api.example.com/v1')).toBe('https://api.example.com/v1')
  })

  it('falls back to the current page origin when the setting is empty', () => {
    expect(resolvePublicApiBaseUrl('')).toBe('https://app.example.com')
    expect(resolvePublicApiBaseUrl('   ')).toBe('https://app.example.com')
    expect(resolvePublicApiBaseUrl(null)).toBe('https://app.example.com')
    expect(resolvePublicApiBaseUrl(undefined)).toBe('https://app.example.com')
  })

  it('does not append /v1 when falling back to the site origin', () => {
    expect(resolvePublicApiBaseUrl('')).toBe('https://app.example.com')
    expect(resolvePublicApiBaseUrl('')).not.toContain('/v1')
  })

  it('accepts an explicit origin override and trims its trailing slash', () => {
    expect(resolvePublicApiBaseUrl('', 'https://origin.example.com/')).toBe('https://origin.example.com')
  })

  it('returns an empty string when neither a config nor an origin is available', () => {
    expect(resolvePublicApiBaseUrl('  ', '')).toBe('')
  })
})
