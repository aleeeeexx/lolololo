import type { PlaygroundGroup } from '@/api/playground'

export type PlaygroundTab = 'chat' | 'image'

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

export function inferPlaygroundModelKinds(name: string): string[] {
  let bare = name.trim().toLowerCase()
  if (bare.startsWith('models/')) bare = bare.slice('models/'.length)
  const slash = bare.lastIndexOf('/')
  if (slash >= 0) {
    bare = bare.slice(slash + 1).trim()
    if (bare.startsWith('models/')) bare = bare.slice('models/'.length)
  }
  const image =
    bare.startsWith('gpt-image-') ||
    bare === 'gpt-image' ||
    bare === 'dall-e' ||
    bare.startsWith('dall-e-') ||
    bare === 'grok-imagine' ||
    bare === 'grok-imagine-edit' ||
    bare.startsWith('grok-imagine-image') ||
    (bare.startsWith('gemini-') && (bare.endsWith('-image') || bare.includes('-image-'))) ||
    bare.startsWith('imagen') ||
    bare.startsWith('imagegeneration')
  const video =
    bare.startsWith('sora') ||
    bare.startsWith('veo') ||
    bare.startsWith('grok-video') ||
    bare.startsWith('grok-imagine-video')
  if (!image && !video) return ['text']
  const kinds: string[] = []
  if (image) kinds.push('image')
  if (video) kinds.push('video')
  return kinds
}

export function playgroundModelKinds(group: PlaygroundGroup | undefined, name: string): string[] {
  const listed = group?.model_kinds?.[name]
  if (listed && listed.length) return listed
  return inferPlaygroundModelKinds(name)
}

export function modelsForPlaygroundGroup(
  group: PlaygroundGroup | undefined,
  kind: PlaygroundTab
): string[] {
  const names = group?.models ?? []
  if (kind === 'image') {
    if (!group?.supports_image) return []
    return names.filter((name) => playgroundModelKinds(group, name).includes('image'))
  }
  return names.filter((name) => {
    const kinds = playgroundModelKinds(group, name)
    return kinds.includes('text') && !kinds.includes('image') && !kinds.includes('video')
  })
}

export function retainOrPickPlaygroundModel(models: string[], current: string): string {
  const want = current.trim()
  if (want && models.includes(want)) return want
  return models[0] ?? ''
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
