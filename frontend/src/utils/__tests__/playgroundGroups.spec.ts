import { describe, expect, it } from 'vitest'
import type { PlaygroundGroup } from '@/api/playground'
import {
  appendPlaygroundAttachmentText,
  inferPlaygroundModelKinds,
  modelsForPlaygroundGroup,
  pickDefaultPlaygroundGroup,
  retainOrPickPlaygroundModel,
} from '@/utils/playgroundGroups'

function group(partial: Partial<PlaygroundGroup> & { id: number; name: string }): PlaygroundGroup {
  return {
    platform: 'openai',
    rate_multiplier: 1,
    subscription_type: false,
    allow_image_generation: true,
    supports_image: true,
    has_remaining_quota: false,
    models: ['gpt-4o'],
    ...partial,
  }
}

describe('playground group selection', () => {
  it('lists chat or image models for the selected group only', () => {
    const chatOnly = group({
      id: 1,
      name: 'chat-only',
      models: ['gpt-4o', 'dall-e-3'],
      supports_image: false,
      model_kinds: { 'gpt-4o': ['text'], 'dall-e-3': ['image'] },
    })
    const image = group({
      id: 2,
      name: 'image',
      models: ['dall-e-3', 'gpt-4o', 'sora-2'],
      supports_image: true,
      model_kinds: { 'dall-e-3': ['image'], 'gpt-4o': ['text'], 'sora-2': ['video'] },
    })
    expect(modelsForPlaygroundGroup(chatOnly, 'chat')).toEqual(['gpt-4o'])
    expect(modelsForPlaygroundGroup(chatOnly, 'image')).toEqual([])
    expect(modelsForPlaygroundGroup(image, 'chat')).toEqual(['gpt-4o'])
    expect(modelsForPlaygroundGroup(image, 'image')).toEqual(['dall-e-3'])
  })

  it('falls back to name-based kinds when session metadata is missing', () => {
    expect(inferPlaygroundModelKinds('dall-e-3')).toEqual(['image'])
    expect(inferPlaygroundModelKinds('gpt-image-1')).toEqual(['image'])
    expect(inferPlaygroundModelKinds('gemini-2.5-flash-image')).toEqual(['image'])
    expect(inferPlaygroundModelKinds('sora-2')).toEqual(['video'])
    expect(inferPlaygroundModelKinds('gpt-4o')).toEqual(['text'])
    const groupWithoutKinds = group({
      id: 3,
      name: 'mixed',
      models: ['gpt-4o', 'dall-e-3'],
      supports_image: true,
    })
    expect(modelsForPlaygroundGroup(groupWithoutKinds, 'chat')).toEqual(['gpt-4o'])
    expect(modelsForPlaygroundGroup(groupWithoutKinds, 'image')).toEqual(['dall-e-3'])
  })

  it('keeps the current model when it is still in the group, otherwise picks the first', () => {
    expect(retainOrPickPlaygroundModel(['gpt-4o', 'gpt-4o-mini'], 'gpt-4o-mini')).toBe('gpt-4o-mini')
    expect(retainOrPickPlaygroundModel(['gpt-4o', 'gpt-4o-mini'], 'claude-sonnet-4')).toBe('gpt-4o')
    expect(retainOrPickPlaygroundModel([], 'gpt-4o')).toBe('')
  })

  it('prefers a subscription with remaining quota, then a lower multiplier', () => {
    const groups = [
      group({ id: 4, name: 'expensive', rate_multiplier: 0.5, subscription_type: false }),
      group({
        id: 2,
        name: 'exhausted-sub',
        rate_multiplier: 0.2,
        subscription_type: true,
        has_remaining_quota: false,
      }),
      group({
        id: 3,
        name: 'active-sub-higher-rate',
        rate_multiplier: 1,
        subscription_type: true,
        has_remaining_quota: true,
      }),
      group({
        id: 1,
        name: 'active-sub-lower-rate',
        rate_multiplier: 0.8,
        subscription_type: true,
        has_remaining_quota: true,
      }),
    ]
    expect(pickDefaultPlaygroundGroup(groups)?.id).toBe(1)
  })

  it('appends attachment text to the user message for follow-up turns', () => {
    expect(appendPlaygroundAttachmentText('summarize this', [
      { filename: 'notes.md', text: '# Hello' },
      { filename: 'scan.pdf' },
    ])).toContain('# Hello')
    expect(appendPlaygroundAttachmentText('summarize this', [
      { filename: 'notes.md', text: '# Hello' },
    ])).toContain('summarize this')
    expect(appendPlaygroundAttachmentText('summarize this', [{ filename: 'scan.pdf' }])).toBe('summarize this')
  })
})
