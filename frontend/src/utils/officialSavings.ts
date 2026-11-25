/**
 * Savings vs official API pricing, using the mid-station's official baseline
 * of 6.8 CNY per 1 USD. A group 倍率 of 1.0x is therefore about 85% cheaper
 * than official: (1 - 倍率 / 6.8) * 100%.
 *
 * When several rates or groups are summarized, the headline uses the lowest
 * applicable 倍率 (the group that is cheapest versus official), which yields
 * the largest savings percent.
 */

export const OFFICIAL_USD_CNY_RATE = 6.8

export interface OfficialSavingsGroupInput {
  name?: string
  rate_multiplier?: number
  user_rate_multiplier?: number | null
  image_rate_independent?: boolean
  image_rate_multiplier?: number
  video_rate_independent?: boolean
  video_rate_multiplier?: number
}

export interface OfficialSavingsGroupSummary {
  name: string
  /** Lowest applicable billing 倍率 for this group (cheapest vs official). */
  rate: number
  /** Rounded savings percent, or null when this group is not cheaper than official. */
  percent: number | null
}

function isFiniteNumber(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value)
}

/** Rounded "cheaper than official" percent, or null when not cheaper. */
export function savingsPercentVsOfficial(
  rate: number,
  officialRate: number = OFFICIAL_USD_CNY_RATE,
): number | null {
  if (!isFiniteNumber(rate) || !isFiniteNumber(officialRate) || officialRate <= 0) return null
  if (rate >= officialRate) return null
  const percent = Math.round((1 - rate / officialRate) * 100)
  return percent > 0 ? percent : null
}

/** Billing multipliers that can apply for a group (default/user + independent image/video). */
export function billingRatesForGroup(group: OfficialSavingsGroupInput): number[] {
  const base = isFiniteNumber(group.user_rate_multiplier)
    ? group.user_rate_multiplier
    : group.rate_multiplier
  const rates: number[] = []
  if (isFiniteNumber(base)) rates.push(base)
  if (group.image_rate_independent && isFiniteNumber(group.image_rate_multiplier)) {
    rates.push(group.image_rate_multiplier)
  }
  if (group.video_rate_independent && isFiniteNumber(group.video_rate_multiplier)) {
    rates.push(group.video_rate_multiplier)
  }
  return rates
}

/** Lowest (cheapest versus official) finite rate, or null. */
export function lowestRate(rates: number[]): number | null {
  const valid = rates.filter((rate) => isFiniteNumber(rate))
  if (valid.length === 0) return null
  return Math.min(...valid)
}

/** Savings for the lowest 倍率 among the given rates (the largest savings percent). */
export function bestSavingsPercentVsOfficial(
  rates: number[],
  officialRate: number = OFFICIAL_USD_CNY_RATE,
): number | null {
  const rate = lowestRate(rates)
  if (rate == null) return null
  return savingsPercentVsOfficial(rate, officialRate)
}

export function summarizeGroupOfficialSavings(
  group: OfficialSavingsGroupInput,
  officialRate: number = OFFICIAL_USD_CNY_RATE,
): OfficialSavingsGroupSummary | null {
  const rate = lowestRate(billingRatesForGroup(group))
  if (rate == null) return null
  const name = typeof group.name === 'string' ? group.name.trim() : ''
  return {
    name,
    rate,
    percent: savingsPercentVsOfficial(rate, officialRate),
  }
}

export function summarizeOfficialSavings(
  groups: OfficialSavingsGroupInput[],
  officialRate: number = OFFICIAL_USD_CNY_RATE,
): { groups: OfficialSavingsGroupSummary[]; percent: number | null; usesMinimum: boolean } {
  const summaries = groups
    .map((group) => summarizeGroupOfficialSavings(group, officialRate))
    .filter((item): item is OfficialSavingsGroupSummary => item != null)
  return {
    groups: summaries,
    percent: bestSavingsPercentVsOfficial(
      summaries.map((item) => item.rate),
      officialRate,
    ),
    usesMinimum: summaries.length > 1,
  }
}
