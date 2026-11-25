export interface PlaygroundHistoryMessage {
  role: 'user' | 'assistant'
  content: string
}

export interface PlaygroundHistoryConversation {
  id: string
  title: string
  model: string
  messages: PlaygroundHistoryMessage[]
  updatedAt: number
}

const MAX_CONVERSATIONS = 20
const MAX_MESSAGES = 80
const MAX_CONTENT = 8000

export function playgroundHistoryKey(userId: number | string): string {
  return `dubdub.playground.history.v1.${userId}`
}

export function loadPlaygroundHistory(userId: number | string): PlaygroundHistoryConversation[] {
  if (typeof localStorage === 'undefined') return []
  try {
    const raw = localStorage.getItem(playgroundHistoryKey(userId))
    if (!raw) return []
    const parsed = JSON.parse(raw) as PlaygroundHistoryConversation[]
    if (!Array.isArray(parsed)) return []
    return parsed
      .filter((item) => item && typeof item.id === 'string' && Array.isArray(item.messages))
      .slice(0, MAX_CONVERSATIONS)
  } catch {
    return []
  }
}

export function savePlaygroundHistory(
  userId: number | string,
  conversations: PlaygroundHistoryConversation[]
): void {
  if (typeof localStorage === 'undefined') return
  const normalized = conversations.slice(0, MAX_CONVERSATIONS).map((item) => ({
    ...item,
    title: String(item.title || 'Chat').slice(0, 80),
    messages: item.messages.slice(-MAX_MESSAGES).map((message) => ({
      role: message.role,
      content: String(message.content || '').slice(0, MAX_CONTENT),
    })),
  }))
  localStorage.setItem(playgroundHistoryKey(userId), JSON.stringify(normalized))
}

export function upsertPlaygroundConversation(
  conversations: PlaygroundHistoryConversation[],
  conversation: PlaygroundHistoryConversation
): PlaygroundHistoryConversation[] {
  const next = conversations.filter((item) => item.id !== conversation.id)
  return [conversation, ...next].slice(0, MAX_CONVERSATIONS)
}
