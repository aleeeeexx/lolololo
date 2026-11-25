<template>
  <div
    v-if="hasContent"
    :class="rootClass"
    data-testid="official-savings"
  >
    <p :class="titleClass" data-testid="official-savings-title">
      {{ titleText }}
    </p>
    <p :class="hintClass" data-testid="official-savings-hint">
      {{ hintText }}
    </p>
    <ul
      v-if="showGroups && summary.groups.length > 0"
      class="mt-3 space-y-1"
      data-testid="official-savings-groups"
    >
      <li
        v-for="(group, index) in summary.groups"
        :key="`${group.name}-${index}`"
        class="text-[13px] text-[color:var(--foreground-secondary)]"
        :data-testid="`official-savings-group-${index}`"
      >
        {{ groupLine(group) }}
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { formatMultiplier } from '@/utils/formatters'
import {
  summarizeOfficialSavings,
  type OfficialSavingsGroupInput,
  type OfficialSavingsGroupSummary,
} from '@/utils/officialSavings'

const props = withDefaults(
  defineProps<{
    groups: OfficialSavingsGroupInput[]
    variant?: 'hero' | 'panel'
    showGroups?: boolean
  }>(),
  {
    variant: 'panel',
    showGroups: false,
  },
)

const { t } = useI18n()

const summary = computed(() => summarizeOfficialSavings(props.groups ?? []))

const hasContent = computed(() => summary.value.groups.length > 0)

const titleText = computed(() => {
  const { percent, usesMinimum } = summary.value
  if (percent == null) return t('officialSavings.notCheaper')
  if (usesMinimum) return t('officialSavings.cheaperMinimum', { percent })
  return t('officialSavings.cheaper', { percent })
})

const hintText = computed(() =>
  summary.value.usesMinimum && summary.value.percent != null
    ? t('officialSavings.baselineMinHint')
    : t('officialSavings.baseline'),
)

const rootClass = computed(() =>
  props.variant === 'hero'
    ? 'official-savings-hero'
    : 'rounded-[length:var(--radius)] border border-[color:var(--border)] bg-[color:var(--surface)] px-4 py-3',
)

const titleClass = computed(() =>
  props.variant === 'hero'
    ? 'text-2xl font-semibold tracking-tight text-[color:var(--accent-ink)] md:text-[28px]'
    : 'text-sm font-semibold text-[color:var(--accent-ink)]',
)

const hintClass = computed(
  () => 'mt-1 text-[13px] leading-relaxed text-[color:var(--muted)]',
)

function groupLine(group: OfficialSavingsGroupSummary): string {
  const rate = formatMultiplier(group.rate)
  const name = group.name || '—'
  if (group.percent == null) {
    return t('officialSavings.groupNotCheaper', { name, rate })
  }
  return t('officialSavings.groupCheaper', { name, rate, percent: group.percent })
}
</script>
