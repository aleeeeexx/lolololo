/**
 * Logged-in catalog of request-side model IDs the current user can call.
 * GET /supported-models is independent of model plaza / available-channels flags.
 */

import { apiClient } from './client'

export type SupportedModelKind = 'text' | 'image' | 'video'

export interface SupportedModelItem {
  /** Exact model ID as returned by the API; copy this value unchanged. */
  name: string
  platforms: string[]
  kinds?: SupportedModelKind[]
  company?: string
}

export async function getSupportedModels(): Promise<SupportedModelItem[]> {
  const { data } = await apiClient.get<SupportedModelItem[] | null>('/supported-models')
  return Array.isArray(data) ? data : []
}

export const supportedModelsAPI = { getSupportedModels }
export default supportedModelsAPI
