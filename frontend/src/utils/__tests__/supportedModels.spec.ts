import { describe, expect, it } from 'vitest'
import {
  filterSupportedModels,
  loadSupportedModels,
  parseSupportedModels,
} from '@/utils/supportedModels'

describe('parseSupportedModels', () => {
  it('dedupes exact model IDs, unions platforms, and skips empty names', () => {
    const items = parseSupportedModels([
      { name: 'gpt-4o', platforms: ['openai'] },
      { name: '  ', platforms: ['openai'] },
      { name: 'claude-sonnet-4', platforms: ['anthropic'] },
      { name: 'gpt-4o', platforms: ['azure'] },
    ])
    expect(items).toEqual([
      { name: 'claude-sonnet-4', platforms: ['anthropic'] },
      { name: 'gpt-4o', platforms: ['azure', 'openai'] },
    ])
  })

  it('treats a missing or non-array payload as an empty list', () => {
    expect(parseSupportedModels(null)).toEqual([])
    expect(parseSupportedModels(undefined)).toEqual([])
    expect(parseSupportedModels({ name: 'gpt-4o' })).toEqual([])
  })
})

describe('filterSupportedModels', () => {
  const models = [
    { name: 'gpt-4o', platforms: ['openai'] },
    { name: 'claude-sonnet-4', platforms: ['anthropic'] },
  ]

  it('filters by model ID or platform, and returns all items for an empty query', () => {
    expect(filterSupportedModels(models, '')).toEqual(models)
    expect(filterSupportedModels(models, 'SONNET')).toEqual([models[1]])
    expect(filterSupportedModels(models, 'openai')).toEqual([models[0]])
  })
})

describe('loadSupportedModels', () => {
  it('parses the session API payload', async () => {
    const models = await loadSupportedModels({
      getSupportedModels: async () => [{ name: 'gpt-4o', platforms: ['openai'] }],
    })
    expect(models).toEqual([{ name: 'gpt-4o', platforms: ['openai'] }])
  })

  it('propagates API errors so the view can show a failure state', async () => {
    await expect(loadSupportedModels({
      getSupportedModels: async () => {
        throw Object.assign(new Error('unauthorized'), { status: 401 })
      },
    })).rejects.toMatchObject({ status: 401 })
  })
})
