import type { PlaygroundGroup } from '@/api/playground'

export type PlaygroundTab = 'chat' | 'image'

export function groupsForPlaygroundModel(
  groups: PlaygroundGroup[] | undefined,
  model: string,
  kind: PlaygroundTab
): PlaygroundGroup[] {
  const list = groups ?? []
  const want = model.trim()
  return list.filter((group) => {
    if (kind === 'image' && !group.supports_image) {
      return false
    }
    if (want && group.models?.length && !group.models.includes(want)) {
      return false
    }
    return true
  })
}

export function pickDefaultPlaygroundGroup(groups: PlaygroundGroup[]): PlaygroundGroup | undefined {
  if (groups.length === 0) return undefined
  return [...groups].sort((a, b) => {
    const aPreferred = a.subscription_type && a.has_remaining_quota ? 0 : 1
    const bPreferred = b.subscription_type && b.has_remaining_quota ? 0 : 1
    if (aPreferred !== bPreferred) return aPreferred - bPreferred
    if (a.rate_multiplier !== b.rate_multiplier) return a.rate_multiplier - b.rate_multiplier
    return a.id - b.id
  })[0]
}

export function appendPlaygroundAttachmentText(
  content: string,
  documents: Array<{ filename: string; text?: string }>
): string {
  const parts = documents
    .filter((doc) => (doc.text || '').trim())
    .map((doc) => `### ${doc.filename}\n${(doc.text || '').trim()}`)
  if (parts.length === 0) return content
  const block = `Attached documents:\n\n${parts.join('\n\n')}`
  const trimmed = content.trim()
  return trimmed ? `${trimmed}\n\n${block}` : block
}
