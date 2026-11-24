/**
 * Helpers for the user-facing "Supported Models" list.
 *
 * Data comes from GET /supported-models, which aggregates request-side model
 * IDs (the same口径 as /v1/models) across groups the current user can bind.
 * The list is not a static all-platform catalog and does not depend on
 * model_plaza_enabled or available_channels_enabled.
 */

export interface SupportedModelItem {
  /** Exact model ID as returned by the API; copy this value unchanged. */
  name: string
  platforms: string[]
}

function normalizeName(name: unknown): string | null {
  if (typeof name !== 'string') return null
  const trimmed = name.trim()
  return trimmed ? trimmed : null
}

function addModel(
  bucket: Map<string, Set<string>>,
  name: unknown,
  platforms: unknown,
): void {
  const modelName = normalizeName(name)
  if (!modelName) return
  let set = bucket.get(modelName)
  if (!set) {
    set = new Set()
    bucket.set(modelName, set)
  }
  const list = Array.isArray(platforms) ? platforms : []
  for (const platform of list) {
    if (typeof platform === 'string' && platform.trim()) {
      set.add(platform.trim())
    }
  }
}

function toSortedItems(bucket: Map<string, Set<string>>): SupportedModelItem[] {
  return [...bucket.entries()]
    .map(([name, platforms]) => ({
      name,
      platforms: [...platforms].sort((a, b) => a.localeCompare(b)),
    }))
    .sort((a, b) => a.name.localeCompare(b.name, undefined, { numeric: true, sensitivity: 'base' }))
}

/** Normalize and dedupe a /supported-models payload. */
export function parseSupportedModels(data: unknown): SupportedModelItem[] {
  const bucket = new Map<string, Set<string>>()
  const rows = Array.isArray(data) ? data : []
  for (const row of rows) {
    if (!row || typeof row !== 'object') continue
    const item = row as { name?: unknown; platforms?: unknown }
    addModel(bucket, item.name, item.platforms)
  }
  return toSortedItems(bucket)
}

export function filterSupportedModels(models: SupportedModelItem[], query: string): SupportedModelItem[] {
  const q = query.trim().toLowerCase()
  if (!q) return models
  return models.filter(
    (model) =>
      model.name.toLowerCase().includes(q) ||
      model.platforms.some((platform) => platform.toLowerCase().includes(q)),
  )
}

export interface LoadSupportedModelsDeps {
  getSupportedModels: () => Promise<unknown>
}

/**
 * Load the supported-model list from the session API.
 * An empty array is a valid empty state, not an error.
 */
export async function loadSupportedModels(
  deps: LoadSupportedModelsDeps,
): Promise<SupportedModelItem[]> {
  return parseSupportedModels(await deps.getSupportedModels())
}
