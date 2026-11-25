/**
 * JWT playground APIs for the logged-in 使用 page.
 * Chat and image generation go through /api/v1/playground/* and never
 * send a user API key to /v1.
 */

import { apiClient } from './client'
import { buildApiUrl } from './url'
import { getLocale } from '@/i18n'

export interface PlaygroundSubscription {
  id: number
  group_id: number
  group_name: string
  status: string
  daily_used_usd: number
  daily_limit_usd: number
  weekly_used_usd: number
  weekly_limit_usd: number
  monthly_used_usd: number
  monthly_limit_usd: number
}

export interface PlaygroundGroup {
  id: number
  name: string
  platform: string
  rate_multiplier: number
  subscription_type: boolean
  allow_image_generation: boolean
  supports_image: boolean
  has_remaining_quota: boolean
  models: string[]
  model_kinds?: Record<string, string[]>
}

export interface PlaygroundSession {
  balance: number
  frozen_balance: number
  available_balance: number
  simple_mode: boolean
  subscriptions: PlaygroundSubscription[]
  groups?: PlaygroundGroup[]
  limits: {
    max_documents: number
    max_document_bytes: number
    max_document_runes: number
  }
}

export interface PlaygroundRequestCost {
  found: boolean
  request_id: string
  model?: string
  actual_cost: number
  total_cost: number
}

export interface PlaygroundChatMessage {
  role: 'system' | 'user' | 'assistant'
  content: string
}

export interface PlaygroundDocumentPayload {
  filename: string
  text?: string
  data?: string
}

export interface PlaygroundImageTask {
  id: string
  task_id: string
  object?: string
  status: 'processing' | 'completed' | 'failed' | string
  http_status?: number
  image_url?: string
  result?: PlaygroundImageResult | null
  error?: { message?: string; type?: string; code?: string } | string | null
  created_at?: number
  completed_at?: number | null
  expires_at?: number
}

export interface PlaygroundImageResult {
  data?: Array<{
    url?: string
    b64_json?: string
  }>
}

function authHeaders(extra?: HeadersInit): Headers {
  const headers = new Headers(extra)
  const token = localStorage.getItem('auth_token')
  if (token && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  headers.set('Accept-Language', getLocale())
  return headers
}

function requestIdFrom(response: Response): string {
  return (
    response.headers.get('X-Client-Request-ID') ||
    response.headers.get('X-Request-Id') ||
    ''
  ).trim()
}

function playgroundErrorMessage(payload: unknown, fallback: string): string {
  if (!payload || typeof payload !== 'object') return fallback
  const record = payload as Record<string, unknown>
  const error = record.error
  if (error && typeof error === 'object') {
    const message = (error as Record<string, unknown>).message
    if (typeof message === 'string' && message.trim()) return message
  }
  if (typeof record.message === 'string' && record.message.trim()) return record.message
  return fallback
}

export async function getPlaygroundSession(): Promise<PlaygroundSession> {
  const { data } = await apiClient.get<PlaygroundSession>('/playground/session', {
    headers: {
      'Cache-Control': 'no-cache',
      Pragma: 'no-cache',
    },
  })
  return data
}

export async function getPlaygroundRequestCost(requestId: string): Promise<PlaygroundRequestCost> {
  const { data } = await apiClient.get<PlaygroundRequestCost>(
    `/playground/requests/${encodeURIComponent(requestId)}`
  )
  return data
}

export async function pollPlaygroundRequestCost(
  requestId: string,
  options?: { attempts?: number; signal?: AbortSignal }
): Promise<PlaygroundRequestCost> {
  const attempts = options?.attempts ?? 6
  let last: PlaygroundRequestCost = {
    found: false,
    request_id: requestId,
    actual_cost: 0,
    total_cost: 0,
  }
  for (let i = 0; i < attempts; i++) {
    if (options?.signal?.aborted) {
      return last
    }
    last = await getPlaygroundRequestCost(requestId)
    if (last.found) return last
    await new Promise((resolve) => setTimeout(resolve, 400 * (i + 1)))
  }
  return last
}

export async function submitPlaygroundChat(input: {
  model: string
  groupId: number
  messages: PlaygroundChatMessage[]
  documents?: PlaygroundDocumentPayload[]
  signal?: AbortSignal
  onDelta: (text: string) => void
}): Promise<{ requestId: string; text: string }> {
  const response = await fetch(buildApiUrl('/playground/chat/completions'), {
    method: 'POST',
    credentials: 'include',
    signal: input.signal,
    headers: authHeaders({
      'Content-Type': 'application/json',
      Accept: 'text/event-stream',
    }),
    body: JSON.stringify({
      model: input.model,
      group_id: input.groupId,
      messages: input.messages,
      stream: true,
      documents: input.documents && input.documents.length > 0 ? input.documents : undefined,
    }),
  })
  const requestId = requestIdFrom(response)
  if (!response.ok || !response.body) {
    let payload: unknown = null
    try {
      payload = await response.json()
    } catch {
      payload = null
    }
    throw new Error(playgroundErrorMessage(payload, `HTTP ${response.status}`))
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let text = ''
  while (true) {
    const { value, done } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    const chunks = buffer.split(/\n\n/)
    buffer = chunks.pop() || ''
    for (const chunk of chunks) {
      const dataLines = chunk
        .split('\n')
        .map((line) => line.replace(/\r$/, ''))
        .filter((line) => line.startsWith('data:'))
        .map((line) => line.slice(5).trimStart())
      if (dataLines.length === 0) continue
      const data = dataLines.join('\n')
      if (!data || data === '[DONE]') continue
      try {
        const parsed = JSON.parse(data) as {
          error?: { message?: string }
          choices?: Array<{ delta?: { content?: string }; message?: { content?: string } }>
        }
        if (parsed.error?.message) {
          throw new Error(parsed.error.message)
        }
        const delta = parsed.choices?.[0]?.delta?.content || parsed.choices?.[0]?.message?.content || ''
        if (delta) {
          text += delta
          input.onDelta(delta)
        }
      } catch (error) {
        if (error instanceof SyntaxError) continue
        throw error
      }
    }
  }
  return { requestId, text }
}

export async function submitPlaygroundImage(input: {
  model: string
  groupId: number
  prompt: string
}): Promise<{ task: PlaygroundImageTask; requestId: string }> {
  const { data, headers } = await apiClient.post<PlaygroundImageTask>(
    '/playground/images/generations',
    { model: input.model, group_id: input.groupId, prompt: input.prompt },
    { timeout: 60000 }
  )
  return {
    task: data,
    requestId: String(headers['x-client-request-id'] || headers['X-Client-Request-ID'] || '').trim(),
  }
}

export async function getPlaygroundImageTask(taskId: string): Promise<PlaygroundImageTask> {
  const { data } = await apiClient.get<PlaygroundImageTask>(
    `/playground/images/tasks/${encodeURIComponent(taskId)}`,
    { timeout: 30000 }
  )
  return data
}

export function playgroundImageContentUrl(taskId: string, index: number): string {
  return buildApiUrl(
    `/playground/images/tasks/${encodeURIComponent(taskId)}/content/${encodeURIComponent(String(index))}`
  )
}

export async function fetchPlaygroundImageContent(taskId: string, index: number): Promise<Blob> {
  const response = await fetch(playgroundImageContentUrl(taskId, index), {
    method: 'GET',
    credentials: 'include',
    headers: authHeaders(),
  })
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }
  return response.blob()
}

export const playgroundAPI = {
  getPlaygroundSession,
  getPlaygroundRequestCost,
  pollPlaygroundRequestCost,
  submitPlaygroundChat,
  submitPlaygroundImage,
  getPlaygroundImageTask,
  playgroundImageContentUrl,
  fetchPlaygroundImageContent,
}

export default playgroundAPI
