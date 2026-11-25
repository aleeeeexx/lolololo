import { beforeEach, describe, expect, it } from 'vitest'
import {
  loadPlaygroundHistory,
  playgroundHistoryKey,
  savePlaygroundHistory,
  upsertPlaygroundConversation,
} from '@/utils/playgroundHistory'

describe('playground history', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('saves and loads conversations for a single user key', () => {
    savePlaygroundHistory(7, [
      {
        id: 'c1',
        title: 'Hello',
        model: 'gpt-4o',
        messages: [{ role: 'user', content: 'hi' }],
        updatedAt: 1,
      },
    ])
    expect(localStorage.getItem(playgroundHistoryKey(7))).toBeTruthy()
    expect(loadPlaygroundHistory(8)).toEqual([])
    expect(loadPlaygroundHistory(7)[0]?.title).toBe('Hello')
  })

  it('upserts a conversation to the front of the list', () => {
    const next = upsertPlaygroundConversation(
      [
        {
          id: 'old',
          title: 'Old',
          model: 'gpt-4o',
          messages: [],
          updatedAt: 1,
        },
      ],
      {
        id: 'new',
        title: 'New',
        model: 'gpt-4o',
        messages: [{ role: 'user', content: 'x' }],
        updatedAt: 2,
      }
    )
    expect(next[0]?.id).toBe('new')
    expect(next).toHaveLength(2)
  })
})
