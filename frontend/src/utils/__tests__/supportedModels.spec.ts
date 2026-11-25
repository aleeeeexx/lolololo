import { describe, expect, it } from 'vitest'
import {
  filterSupportedModels,
  listSupportedModelCompanies,
  loadSupportedModels,
  parseSupportedModels,
} from '@/utils/supportedModels'

const gpt4o = {
  name: 'gpt-4o',
  platforms: ['openai'],
  kinds: ['text'] as const,
  company: 'openai',
}
const claude = {
  name: 'claude-sonnet-4',
  platforms: ['anthropic'],
  kinds: ['text'] as const,
  company: 'anthropic',
}
const dalle = {
  name: 'dall-e-3',
  platforms: ['openai'],
  kinds: ['image'] as const,
  company: 'openai',
}
const sora = {
  name: 'sora-2',
  platforms: ['openai'],
  kinds: ['video'] as const,
  company: 'openai',
}
const grokVideo = {
  name: 'grok-imagine-video',
  platforms: ['grok'],
  kinds: ['video'] as const,
  company: 'xai',
}
const custom = {
  name: 'mystery-custom-model',
  platforms: ['grok'],
  kinds: ['text'] as const,
  company: 'other',
}
const hybrid = {
  name: 'gpt-image-sora-hybrid',
  platforms: ['openai'],
  kinds: ['image', 'video'] as const,
  company: 'openai',
}

describe('parseSupportedModels', () => {
  it('dedupes exact model IDs, unions platforms, and skips empty names', () => {
    const items = parseSupportedModels([
      { name: 'gpt-4o', platforms: ['openai'], kinds: ['text'], company: 'openai' },
      { name: '  ', platforms: ['openai'] },
      { name: 'claude-sonnet-4', platforms: ['anthropic'], kinds: ['text'], company: 'anthropic' },
      { name: 'gpt-4o', platforms: ['azure'], kinds: ['text'], company: 'openai' },
    ])
    expect(items).toEqual([
      { name: 'claude-sonnet-4', platforms: ['anthropic'], kinds: ['text'], company: 'anthropic' },
      { name: 'gpt-4o', platforms: ['azure', 'openai'], kinds: ['text'], company: 'openai' },
    ])
  })

  it('keeps kinds and company, and treats a missing company as other', () => {
    const items = parseSupportedModels([
      { name: 'dall-e-3', platforms: ['openai'], kinds: ['image'], company: 'OpenAI' },
      { name: 'mystery-custom-model', platforms: ['grok'] },
    ])
    expect(items).toEqual([
      { name: 'dall-e-3', platforms: ['openai'], kinds: ['image'], company: 'openai' },
      { name: 'mystery-custom-model', platforms: ['grok'], kinds: [], company: 'other' },
    ])
  })

  it('treats a missing or non-array payload as an empty list', () => {
    expect(parseSupportedModels(null)).toEqual([])
    expect(parseSupportedModels(undefined)).toEqual([])
    expect(parseSupportedModels({ name: 'gpt-4o' })).toEqual([])
  })
})

describe('filterSupportedModels', () => {
  const models = [gpt4o, claude, dalle, sora, grokVideo, custom, hybrid]

  it('filters by model ID or platform, and returns all items for an empty query', () => {
    expect(filterSupportedModels(models, '')).toEqual(models)
    expect(filterSupportedModels(models, 'SONNET')).toEqual([claude])
    expect(filterSupportedModels(models, 'anthropic')).toEqual([claude])
  })

  it('filters by kind without falling back to the full list', () => {
    expect(filterSupportedModels(models, { kind: 'image' }).map((model) => model.name)).toEqual([
      'dall-e-3',
      'gpt-image-sora-hybrid',
    ])
    expect(filterSupportedModels(models, { kind: 'video' }).map((model) => model.name)).toEqual([
      'sora-2',
      'grok-imagine-video',
      'gpt-image-sora-hybrid',
    ])
    expect(filterSupportedModels(models, { kind: 'text' }).map((model) => model.name)).toEqual([
      'gpt-4o',
      'claude-sonnet-4',
      'mystery-custom-model',
    ])
  })

  it('filters by company, including other, without falling back to the full list', () => {
    expect(filterSupportedModels(models, { company: 'xai' }).map((model) => model.name)).toEqual([
      'grok-imagine-video',
    ])
    expect(filterSupportedModels(models, { company: 'other' }).map((model) => model.name)).toEqual([
      'mystery-custom-model',
    ])
    expect(filterSupportedModels(models, { company: 'mistral' })).toEqual([])
  })

  it('combines kind, company, and search with AND, and stays empty when nothing matches', () => {
    expect(filterSupportedModels(models, { kind: 'video', company: 'openai' }).map((model) => model.name)).toEqual([
      'sora-2',
      'gpt-image-sora-hybrid',
    ])
    expect(filterSupportedModels(models, { kind: 'image', company: 'openai', query: 'dall' }).map((model) => model.name)).toEqual([
      'dall-e-3',
    ])
    expect(filterSupportedModels(models, { kind: 'image', company: 'anthropic' })).toEqual([])
    expect(filterSupportedModels(models, { kind: 'text', query: 'sora' })).toEqual([])
  })
})

describe('listSupportedModelCompanies', () => {
  it('returns companies that actually appear, with other last', () => {
    expect(listSupportedModelCompanies([custom, grokVideo, gpt4o, claude])).toEqual([
      'openai',
      'anthropic',
      'xai',
      'other',
    ])
  })
})

describe('loadSupportedModels', () => {
  it('parses the session API payload', async () => {
    const models = await loadSupportedModels({
      getSupportedModels: async () => [{ name: 'gpt-4o', platforms: ['openai'], kinds: ['text'], company: 'openai' }],
    })
    expect(models).toEqual([{ name: 'gpt-4o', platforms: ['openai'], kinds: ['text'], company: 'openai' }])
  })

  it('propagates API errors so the view can show a failure state', async () => {
    await expect(loadSupportedModels({
      getSupportedModels: async () => {
        throw Object.assign(new Error('unauthorized'), { status: 401 })
      },
    })).rejects.toMatchObject({ status: 401 })
  })
})
