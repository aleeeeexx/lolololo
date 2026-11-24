<template>
  <div class="min-h-screen bg-[color:var(--background)] text-[color:var(--foreground)]">
    <header class="app-topbar">
      <div class="content-wrap flex h-full items-center justify-between gap-4">
        <RouterLink to="/home" class="flex min-w-0 items-center gap-3">
          <template v-if="settings">
            <span class="brand-mark">
              <img :src="siteLogo || '/logo.svg'" alt="Logo" />
            </span>
            <span class="truncate text-[13px] font-semibold text-[color:var(--foreground)]">
              {{ siteName }}
            </span>
          </template>
          <template v-else>
            <span class="h-10 w-10 flex-shrink-0 animate-pulse rounded-xl bg-gray-200 dark:bg-dark-700" aria-hidden="true"></span>
            <span class="h-5 w-28 animate-pulse rounded bg-gray-200 dark:bg-dark-700" aria-hidden="true"></span>
          </template>
        </RouterLink>
        <RouterLink
          to="/login"
          class="inline-flex flex-shrink-0 items-center justify-center rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white shadow-sm shadow-primary-600/20 transition hover:bg-primary-700"
        >
          {{ t('home.login') }}
        </RouterLink>
      </div>
    </header>

    <main class="content-wrap max-w-4xl py-8 lg:py-10">
      <div v-if="loading" class="flex min-h-[320px] items-center justify-center">
        <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary-600"></div>
      </div>

      <section
        v-else-if="loadError"
        class="rounded-lg border border-red-200 bg-red-50 p-6 text-red-700 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-200"
      >
        <h1 class="text-lg font-semibold">{{ t('legal.loadFailed') }}</h1>
        <p class="mt-2 text-sm">{{ t('legal.retryLater') }}</p>
      </section>

      <section
        v-else-if="!currentDocument"
        class="rounded-lg border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-900"
      >
        <div class="flex items-start gap-3">
          <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-md bg-gray-100 text-gray-600 dark:bg-dark-800 dark:text-dark-300">
            <Icon name="document" size="sm" />
          </span>
          <div>
            <h1 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('legal.notFound') }}</h1>
            <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-dark-300">
              {{ t('legal.notFoundDescription') }}
            </p>
          </div>
        </div>
      </section>

      <article v-else>
        <div class="mb-8 border-b border-gray-200 pb-6 dark:border-dark-700">
          <div class="flex items-start gap-4">
            <span class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-md bg-primary-50 text-primary-700 dark:bg-primary-500/10 dark:text-primary-300">
              <Icon :name="documentIcon" size="md" />
            </span>
            <div class="min-w-0">
              <p class="text-sm font-medium text-primary-700 dark:text-primary-300">{{ documentTypeLabel }}</p>
              <h1 class="mt-2 break-words text-[36px] font-semibold tracking-normal text-[color:var(--foreground)] leading-[1.2]">
                {{ currentDocument.title }}
              </h1>
              <p v-if="updatedAt" class="mt-3 text-sm text-gray-500 dark:text-dark-400">
                {{ t('legal.updatedAt', { date: updatedAt }) }}
              </p>
            </div>
          </div>
        </div>

        <div
          v-if="hasContent"
          class="legal-document-content"
          v-html="renderedHtml"
        ></div>
        <div
          v-else
          class="rounded-lg border border-dashed border-gray-300 bg-white px-6 py-14 text-center text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-400"
        >
          {{ t('legal.empty') }}
        </div>
      </article>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { getLocale } from '@/i18n'
import { sanitizeUrl } from '@/utils/url'
import { useAppStore } from '@/stores/app'
import type { LoginAgreementDocument } from '@/types'
import zhAdminCompliance from '../../../../docs/legal/admin-compliance.zh.md?raw'
import enAdminCompliance from '../../../../docs/legal/admin-compliance.en.md?raw'
import { DEFAULT_SITE_NAME } from '@/constants/branding'

type LegalDocumentIcon = 'document' | 'shield' | 'globe' | 'cog'

const route = useRoute()
const { t } = useI18n()
const appStore = useAppStore()
const settings = computed(() => appStore.cachedPublicSettings)
const loading = ref(!settings.value)
const loadError = ref(false)

marked.setOptions({
  breaks: true,
  gfm: true,
})

const documentId = computed(() => String(route.params.documentId || ''))
const isAdminComplianceDocument = computed(() => documentId.value === 'admin-compliance')
const documents = computed(() => settings.value?.login_agreement_documents ?? [])
const siteName = computed(() => settings.value?.site_name || DEFAULT_SITE_NAME)
const siteLogo = computed(() => sanitizeUrl(settings.value?.site_logo || '', {
  allowRelative: true,
  allowDataUrl: true,
}))
const updatedAt = computed(() =>
  isAdminComplianceDocument.value ? '' : settings.value?.login_agreement_updated_at || ''
)
const documentTypeLabel = computed(() =>
  isAdminComplianceDocument.value ? t('legal.adminCompliance') : t('legal.loginAgreement')
)

const currentDocument = computed<LoginAgreementDocument | null>(() => {
  if (isAdminComplianceDocument.value) {
    return {
      id: 'admin-compliance',
      title: t('adminCompliance.title'),
      content_md: getLocale() === 'zh' ? zhAdminCompliance : enAdminCompliance
    }
  }
  const id = documentId.value
  if (!id) {
    return null
  }
  return documents.value.find((doc) => doc.id === id) ?? null
})

const hasContent = computed(() => Boolean(currentDocument.value?.content_md?.trim()))

const renderedHtml = computed(() => {
  const content = currentDocument.value?.content_md?.trim() || ''
  if (!content) {
    return ''
  }
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
})

const documentIcon = computed<LegalDocumentIcon>(() => {
  const title = currentDocument.value?.title || ''
  if (title.includes('政策') || title.includes('隐私')) {
    return 'shield'
  }
  if (title.includes('国家') || title.includes('地区')) {
    return 'globe'
  }
  if (title.includes('特定')) {
    return 'cog'
  }
  return 'document'
})

onMounted(async () => {
  loadError.value = false
  const loadedSettings = await appStore.fetchPublicSettings()
  if (!loadedSettings) {
    loadError.value = true
  }
  loading.value = false
})
</script>

<style scoped>
.legal-document-content {
  font-size: 17px;
  line-height: 1.82;
  overflow-wrap: anywhere;
  color: var(--foreground-secondary);
}

.legal-document-content :deep(h1),
.legal-document-content :deep(h2),
.legal-document-content :deep(h3),
.legal-document-content :deep(h4) {
  color: var(--foreground);
  font-weight: 600;
  letter-spacing: 0;
}

.legal-document-content :deep(h1) {
  margin: 2rem 0 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--border);
  font-size: 36px;
  line-height: 1.2;
}

.legal-document-content :deep(h2) {
  margin: 1.75rem 0 0.75rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--border);
  font-size: 24px;
}

.legal-document-content :deep(h3) {
  margin: 1.5rem 0 0.5rem;
  font-size: 1.25rem;
}

.legal-document-content :deep(h4) {
  margin: 1.25rem 0 0.5rem;
  font-size: 1.125rem;
}

.legal-document-content :deep(p) {
  margin-bottom: 1rem;
  color: var(--foreground-secondary);
}

.legal-document-content :deep(a) {
  color: var(--accent-ink);
  text-decoration: underline;
  text-underline-offset: 4px;
}

.legal-document-content :deep(a:hover) {
  color: var(--accent-hover);
}

.legal-document-content :deep(ul) {
  margin-bottom: 1rem;
  list-style: disc;
  padding-left: 1.5rem;
}

.legal-document-content :deep(ol) {
  margin-bottom: 1rem;
  list-style: decimal;
  padding-left: 1.5rem;
}

.legal-document-content :deep(li) {
  margin-bottom: 0.25rem;
  color: var(--foreground-secondary);
}

.legal-document-content :deep(blockquote) {
  margin: 1.25rem 0;
  padding: 12px 16px;
  color: var(--foreground-secondary);
  background: var(--surface);
  border: 1px solid var(--border);
  border-left: 3px solid var(--accent-ink);
  border-radius: var(--radius-sm);
}

.legal-document-content :deep(code) {
  border-radius: var(--radius-sm);
  background: var(--surface-deep);
  padding: 2px 6px;
  font-family: var(--font-mono), ui-monospace, monospace;
  font-size: 0.875rem;
  color: var(--foreground);
}

.legal-document-content :deep(pre) {
  margin: 1.25rem 0;
  overflow-x: auto;
  border-radius: var(--radius);
  background: #0c0c0c;
  padding: 16px;
  color: #cccccc;
  border: 1px solid var(--border);
}

.legal-document-content :deep(pre code) {
  background: transparent;
  padding: 0;
  color: inherit;
}

.legal-document-content :deep(table) {
  margin: 1.25rem 0;
  display: block;
  width: 100%;
  overflow-x: auto;
  border-collapse: collapse;
}

.legal-document-content :deep(th),
.legal-document-content :deep(td) {
  border: 1px solid var(--border);
  padding: 8px 12px;
}

.legal-document-content :deep(th) {
  background: var(--surface-deep);
  text-align: left;
  font-weight: 600;
  color: var(--foreground);
}

.legal-document-content :deep(img) {
  margin: 1.25rem 0;
  height: auto;
  max-width: 100%;
  border-radius: var(--radius);
  border: 1px solid var(--border);
}

.legal-document-content :deep(hr) {
  margin: 1.75rem 0;
  border-color: var(--border);
}
</style>
