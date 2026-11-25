import { describe, expect, it } from 'vitest'
import {
  OFFICIAL_USD_CNY_RATE,
  bestSavingsPercentVsOfficial,
  billingRatesForGroup,
  lowestRate,
  savingsPercentVsOfficial,
  summarizeGroupOfficialSavings,
  summarizeOfficialSavings,
} from '@/utils/officialSavings'

describe('officialSavings', () => {
  it('uses the official 6.8 CNY per USD baseline', () => {
    expect(OFFICIAL_USD_CNY_RATE).toBe(6.8)
  })

  it('matches the posted examples: 1.0x ≈ 85% and 0.5x ≈ 93%', () => {
    expect(savingsPercentVsOfficial(1)).toBe(85)
    expect(savingsPercentVsOfficial(0.5)).toBe(93)
    expect(savingsPercentVsOfficial(0.1)).toBe(99)
  })

  it('rounds the production cheapest public group 0.35x to 95%', () => {
    // 1 - 0.35 / 6.8 = 0.94852941… → 94.8529…% → 95%
    expect(savingsPercentVsOfficial(0.35)).toBe(95)
  })

  it('does not claim savings when the rate is at or above official', () => {
    expect(savingsPercentVsOfficial(6.8)).toBeNull()
    expect(savingsPercentVsOfficial(7)).toBeNull()
    expect(savingsPercentVsOfficial(6.79)).toBeNull()
  })

  it('rejects invalid rates', () => {
    expect(savingsPercentVsOfficial(Number.NaN)).toBeNull()
    expect(savingsPercentVsOfficial(Number.POSITIVE_INFINITY)).toBeNull()
    expect(savingsPercentVsOfficial(1, 0)).toBeNull()
  })

  it('treats a zero 倍率 as 100% cheaper', () => {
    expect(savingsPercentVsOfficial(0)).toBe(100)
  })

  it('includes independent image and video rates when those switches are on', () => {
    expect(
      billingRatesForGroup({
        rate_multiplier: 0.5,
        user_rate_multiplier: 0.8,
        image_rate_independent: true,
        image_rate_multiplier: 1.2,
        video_rate_independent: true,
        video_rate_multiplier: 0.2,
      }),
    ).toEqual([0.8, 1.2, 0.2])

    expect(
      billingRatesForGroup({
        rate_multiplier: 1,
        image_rate_independent: false,
        image_rate_multiplier: 9,
        video_rate_independent: false,
        video_rate_multiplier: 9,
      }),
    ).toEqual([1])
  })

  it('uses the lowest 倍率 so the headline is for the group cheapest versus official', () => {
    expect(lowestRate([0.5, 1, 0.2])).toBe(0.2)
    expect(bestSavingsPercentVsOfficial([0.5, 1])).toBe(93)
    expect(bestSavingsPercentVsOfficial([0.5])).toBe(93)
    expect(bestSavingsPercentVsOfficial([0.35, 0.5, 0.6, 1, 1.6])).toBe(95)
  })

  it('uses a group\'s lowest applicable 倍率 on that group\'s line', () => {
    const summary = summarizeGroupOfficialSavings({
      name: 'mixed',
      rate_multiplier: 0.5,
      user_rate_multiplier: 0.8,
      image_rate_independent: true,
      image_rate_multiplier: 1.2,
      video_rate_independent: true,
      video_rate_multiplier: 0.2,
    })
    // 1 - 0.2 / 6.8 = 0.97058823… → 97%
    expect(summary).toEqual({ name: 'mixed', rate: 0.2, percent: 97 })
  })

  it('still headlines the cheapest group when another group is not cheaper than official', () => {
    const summary = summarizeOfficialSavings([
      { name: 'cheap', rate_multiplier: 0.5 },
      { name: 'full', rate_multiplier: 7 },
    ])
    expect(summary.usesMinimum).toBe(true)
    expect(summary.percent).toBe(93)
    expect(summary.groups.map((item) => item.percent)).toEqual([93, null])
  })

  it('treats a single cheaper group as an exact figure, not a minimum-across-groups headline', () => {
    const summary = summarizeOfficialSavings([{ name: 'default', rate_multiplier: 1 }])
    expect(summary.usesMinimum).toBe(false)
    expect(summary.percent).toBe(85)
    expect(summary.groups[0]).toEqual({ name: 'default', rate: 1, percent: 85 })
  })
})
