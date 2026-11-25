/**
 * Public, unauthenticated billing multipliers for active non-exclusive groups.
 * Independent of the model-plaza feature flag.
 */

import { apiClient } from './client'

export interface PublicGroupRate {
  id: number
  name: string
  rate_multiplier: number
  image_rate_independent: boolean
  image_rate_multiplier: number
  video_rate_independent: boolean
  video_rate_multiplier: number
}

export interface PublicGroupRatesResponse {
  groups: PublicGroupRate[]
}

export async function getPublicGroupRates(): Promise<PublicGroupRate[]> {
  const { data } = await apiClient.get<PublicGroupRatesResponse | PublicGroupRate[] | null>(
    '/public/group-rates',
  )
  if (Array.isArray(data)) return data
  return Array.isArray(data?.groups) ? data.groups : []
}

export const publicGroupRatesAPI = { getPublicGroupRates }
export default publicGroupRatesAPI
