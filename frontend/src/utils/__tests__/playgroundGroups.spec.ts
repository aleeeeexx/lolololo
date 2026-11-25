import { describe, expect, it } from 'vitest'
import type { PlaygroundGroup } from '@/api/playground'
import {
  appendPlaygroundAttachmentText,
  groupsForPlaygroundModel,
  pickDefaultPlaygroundGroup,
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
  it('filters groups by advertised model and image support', () => {
    const groups = [
      group({ id: 1, name: 'chat-only', models: ['gpt-4o'], supports_image: false }),
      group({ id: 2, name: 'image', models: ['dall-e-3', 'gpt-4o'], supports_image: true }),
      group({ id: 3, name: 'other', models: ['claude-sonnet-4'], supports_image: true }),
    ]
    expect(groupsForPlaygroundModel(groups, 'gpt-4o', 'chat').map((item) => item.id)).toEqual([1, 2])
    expect(groupsForPlaygroundModel(groups, 'gpt-4o', 'image').map((item) => item.id)).toEqual([2])
    expect(groupsForPlaygroundModel(groups, 'claude-sonnet-4', 'chat').map((item) => item.id)).toEqual([3])
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
