<template>
  <AppLayout>
    <div class="space-y-5">
      <div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
        <div class="relative w-full sm:w-80">
          <Icon
            name="search"
            size="md"
            class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
          />
          <input
            v-model="searchQuery"
            type="text"
            data-testid="supported-models-search"
            :placeholder="t('supportedModels.searchPlaceholder')"
            class="input pl-10"
          />
        </div>
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="loading"
          :title="t('common.refresh', 'Refresh')"
          data-testid="supported-models-refresh"
          @click="loadModels"
        >
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>

      <p v-if="!loading && !loadFailed && sourceHint" class="text-xs text-gray-500 dark:text-dark-400" data-testid="supported-models-source">
        {{ sourceHint }}
      </p>

      <div v-if="loading" class="flex min-h-[240px] items-center justify-center" data-testid="supported-models-loading">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-600/25 border-t-primary-600 dark:border-primary-400/25 dark:border-t-primary-400"></div>
      </div>

      <div
        v-else-if="loadFailed"
        class="rounded border border-red-200 bg-red-50 px-5 py-8 text-center text-sm text-red-600 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-300"
        role="alert"
        data-testid="supported-models-error"
      >
        {{ t('supportedModels.loadFailed') }}
      </div>

      <div
        v-else-if="filteredModels.length === 0"
        class="rounded border border-dashed border-gray-300 px-5 py-12 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-dark-400"
        role="status"
        data-testid="supported-models-empty"
      >
        {{ searchActive ? t('supportedModels.noSearchResult') : t('supportedModels.empty') }}
      </div>

      <div v-else class="overflow-hidden rounded border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900" data-testid="supported-models-list">
        <p class="border-b border-gray-100 px-4 py-2 text-xs text-gray-500 dark:border-dark-800 dark:text-dark-400">
          {{ t('supportedModels.count', { count: filteredModels.length }) }}
        </p>
        <ul>
          <li
            v-for="model in filteredModels"
            :key="model.name"
            class="border-b border-gray-100 last:border-b-0 dark:border-dark-800"
          >
            <button
              type="button"
              class="flex w-full items-center justify-between gap-3 px-4 py-3 text-left transition-colors hover:bg-gray-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-primary-500 dark:hover:bg-dark-800"
              :data-testid="`supported-model-${model.name}`"
              :data-model-id="model.name"
              :title="t('supportedModels.copyModelId', { model: model.name })"
              :aria-label="t('supportedModels.copyModelId', { model: model.name })"
              @click="copyModel(model.name)"
            >
              <span class="min-w-0">
                <span class="block break-all font-mono text-sm font-medium text-gray-900 dark:text-white">{{ model.name }}</span>
                <span v-if="model.platforms.length" class="mt-0.5 block text-xs text-gray-500 dark:text-dark-400">
                  {{ model.platforms.join(', ') }}
                </span>
              </span>
              <Icon name="copy" size="sm" class="shrink-0 text-gray-500 dark:text-dark-300" aria-hidden="true" />
            </button>
          </li>
        </ul>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { getSupportedModels } from '@/api/supportedModels'
import { useClipboard } from '@/composables/useClipboard'
import {
  filterSupportedModels,
  loadSupportedModels,
  type SupportedModelItem,
} from '@/utils/supportedModels'

const { t } = useI18n()
const { copyToClipboard } = useClipboard()

const models = ref<SupportedModelItem[]>([])
const loading = ref(false)
const loadFailed = ref(false)
const loaded = ref(false)
const searchQuery = ref('')

const searchActive = computed(() => searchQuery.value.trim() !== '')
const filteredModels = computed(() => filterSupportedModels(models.value, searchQuery.value))

const sourceHint = computed(() => (loaded.value ? t('supportedModels.sourceHint') : ''))

async function loadModels() {
  loading.value = true
  loadFailed.value = false
  loaded.value = false
  try {
    models.value = await loadSupportedModels({ getSupportedModels })
    loaded.value = true
  } catch (err: unknown) {
    console.error('Failed to load supported models:', err)
    loadFailed.value = true
    models.value = []
  } finally {
    loading.value = false
  }
}

async function copyModel(name: string) {
  await copyToClipboard(name, t('supportedModels.copied', { model: name }))
}

onMounted(loadModels)
</script>
