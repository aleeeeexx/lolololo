/**
 * Resolve the public API base URL shown to users for client setup.
 *
 * Product convention (admin "API Base URL" / "API 端点地址"):
 * this is the site endpoint used by "Use Key", CC Switch import, and
 * this keys-page copy control. The admin placeholder is a site origin
 * such as https://api.example.com, not an OpenAI-style /v1 path.
 *
 * Rules:
 * - Prefer the trimmed admin-configured `api_base_url` when present.
 * - Otherwise fall back to the current page origin.
 * - Strip trailing slashes only.
 * - Do not append or invent /v1, /v1beta, or /api/v1. OpenAI-compatible
 *   clients may add /v1 themselves; Anthropic and Codex use different
 *   path conventions. If an admin already included a suffix such as /v1,
 *   it is kept as configured after slash trimming.
 * - Never hardcode a production host or localhost.
 */
export function resolvePublicApiBaseUrl(
  configured?: string | null,
  origin: string = typeof window !== 'undefined' ? window.location.origin : '',
): string {
  const fromConfig = (configured ?? '').trim().replace(/\/+$/, '')
  if (fromConfig) return fromConfig
  return origin.trim().replace(/\/+$/, '')
}
