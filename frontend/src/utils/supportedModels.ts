/**
 * Helpers for the user-facing "Supported Models" list.
 *
 * Data comes from GET /supported-models, which aggregates request-side model
 * IDs (the same口径 as /v1/models) across groups the current user can bind.
 * The list is not a static all-platform catalog and does not depend on
 * model_plaza_enabled or available_channels_enabled.
 */

export const SUPPORTED_MODEL_KINDS = ['text', 'image', 'video'] as const
export type SupportedModelKind = (typeof SUPPORTED_MODEL_KINDS)[number]

export const SUPPORTED_MODEL_COMPANY_OTHER = 'other'

/** Stable vendor order for filter chips; unknown slugs are appended, "other" last. */
export const SUPPORTED_MODEL_COMPANY_ORDER = [
  'openai',
  'anthropic',
  'google',
  'xai',
  'deepseek',
  'moonshot',
  'zhipu',
  'minimax',
  'alibaba',
  'meta',
  'mistral',
  'bytedance',
  'tencent',
  'baidu',
  'baichuan',
] as const

export interface SupportedModelItem {
  /** Exact model ID as returned by the API; copy this value unchanged. */
  name: string
  platforms: string[]
  /** Generation kinds: text, image, and/or video. A model may have more than one. */
  kinds: SupportedModelKind[]
  /** Model vendor slug (openai, google, xai, ...), or "other". Not a gateway platform. */
  company: string
}

export interface SupportedModelFilters {
  query?: string
  /** "all" or empty means no type filter. */
  kind?: SupportedModelKind | 'all' | ''
  /** "all" or empty means no company filter. */
  company?: string
}

const KNOWN_KINDS = new Set<string>(SUPPORTED_MODEL_KINDS)

function normalizeName(name: unknown): string | null {
  if (typeof name !== 'string') return null
  const trimmed = name.trim()
  return trimmed ? trimmed : null
}

function normalizeKind(value: unknown): SupportedModelKind | null {
  if (typeof value !== 'string') return null
  const kind = value.trim().toLowerCase()
  return KNOWN_KINDS.has(kind) ? (kind as SupportedModelKind) : null
}

function normalizeCompany(value: unknown): string {
  if (typeof value !== 'string') return SUPPORTED_MODEL_COMPANY_OTHER
  const company = value.trim().toLowerCase()
  return company || SUPPORTED_MODEL_COMPANY_OTHER
}

function parseKinds(value: unknown): SupportedModelKind[] {
  const list = Array.isArray(value) ? value : []
  const kinds: SupportedModelKind[] = []
  const seen = new Set<SupportedModelKind>()
  for (const kind of SUPPORTED_MODEL_KINDS) {
    // Preserve the canonical text → image → video order even if the payload is shuffled.
    if (list.some((item) => normalizeKind(item) === kind) && !seen.has(kind)) {
      seen.add(kind)
      kinds.push(kind)
    }
  }
  return kinds
}

function parsePlatforms(value: unknown): string[] {
  const list = Array.isArray(value) ? value : []
  const platforms: string[] = []
  const seen = new Set<string>()
  for (const platform of list) {
    if (typeof platform !== 'string') continue
    const trimmed = platform.trim()
    if (!trimmed || seen.has(trimmed)) continue
    seen.add(trimmed)
    platforms.push(trimmed)
  }
  return platforms
}

/** Normalize and dedupe a /supported-models payload. */
export function parseSupportedModels(data: unknown): SupportedModelItem[] {
  const bucket = new Map<string, SupportedModelItem>()
  const rows = Array.isArray(data) ? data : []
  for (const row of rows) {
    if (!row || typeof row !== 'object') continue
    const item = row as { name?: unknown; platforms?: unknown; kinds?: unknown; company?: unknown }
    const modelName = normalizeName(item.name)
    if (!modelName) continue
    const incoming: SupportedModelItem = {
      name: modelName,
      platforms: parsePlatforms(item.platforms),
      kinds: parseKinds(item.kinds),
      company: normalizeCompany(item.company),
    }
    const existing = bucket.get(modelName)
    if (!existing) {
      bucket.set(modelName, incoming)
      continue
    }
    const platformSet = new Set([...existing.platforms, ...incoming.platforms])
    const kindSet = new Set([...existing.kinds, ...incoming.kinds])
    existing.platforms = [...platformSet].sort((a, b) => a.localeCompare(b))
    existing.kinds = SUPPORTED_MODEL_KINDS.filter((kind) => kindSet.has(kind))
    if (existing.company === SUPPORTED_MODEL_COMPANY_OTHER && incoming.company !== SUPPORTED_MODEL_COMPANY_OTHER) {
      existing.company = incoming.company
    }
  }
  return [...bucket.values()].sort((a, b) =>
    a.name.localeCompare(b.name, undefined, { numeric: true, sensitivity: 'base' }),
  )
}

/**
 * Filter the catalog by search text, kind, and company (AND).
 * An unmatched combination returns an empty list and must not fall back to the
 * full catalog. Passing a string is treated as a search query only.
 */
export function filterSupportedModels(
  models: SupportedModelItem[],
  queryOrFilters: string | SupportedModelFilters = '',
): SupportedModelItem[] {
  const filters: SupportedModelFilters =
    typeof queryOrFilters === 'string' ? { query: queryOrFilters } : queryOrFilters ?? {}
  const query = (filters.query ?? '').trim().toLowerCase()
  const kind = typeof filters.kind === 'string' ? filters.kind.trim().toLowerCase() : ''
  const company = typeof filters.company === 'string' ? filters.company.trim().toLowerCase() : ''
  const kindActive = kind !== '' && kind !== 'all'
  const companyActive = company !== '' && company !== 'all'

  return models.filter((model) => {
    if (kindActive && !model.kinds.includes(kind as SupportedModelKind)) return false
    if (companyActive && model.company !== company) return false
    if (!query) return true
    return (
      model.name.toLowerCase().includes(query) ||
      model.platforms.some((platform) => platform.toLowerCase().includes(query)) ||
      model.company.toLowerCase().includes(query)
    )
  })
}

/** Companies that actually appear in the current catalog, in display order. */
export function listSupportedModelCompanies(models: SupportedModelItem[]): string[] {
  const present = new Set(models.map((model) => model.company || SUPPORTED_MODEL_COMPANY_OTHER))
  const preferred = SUPPORTED_MODEL_COMPANY_ORDER.filter((company) => present.has(company))
  const rest = [...present]
    .filter(
      (company) =>
        company !== SUPPORTED_MODEL_COMPANY_OTHER &&
        !SUPPORTED_MODEL_COMPANY_ORDER.includes(company as (typeof SUPPORTED_MODEL_COMPANY_ORDER)[number]),
    )
    .sort((a, b) => a.localeCompare(b))
  const out = [...preferred, ...rest]
  if (present.has(SUPPORTED_MODEL_COMPANY_OTHER)) out.push(SUPPORTED_MODEL_COMPANY_OTHER)
  return out
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
