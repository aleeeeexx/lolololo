<template>
  <div class="handbook-page">
    <aside class="handbook-toc" aria-label="Table of contents">
      <div class="handbook-toc-title">{{ t('handbook.toc') }}</div>
      <nav class="handbook-toc-nav">
        <a
          v-for="item in tocItems"
          :key="item.id"
          :href="`#${item.id}`"
          class="handbook-toc-link"
          :class="{ 'is-active': activeHeadingId === item.id }"
          @click.prevent="scrollToHeading(item.id)"
        >
          {{ item.text }}
        </a>
      </nav>
    </aside>

    <article class="handbook-main">
      <p
        v-if="localeStartsWithEn"
        class="mb-4 rounded-lg border border-[color:var(--border)] bg-[color:var(--surface-deep)] px-4 py-3 text-sm text-[color:var(--foreground-secondary)]"
      >
        {{ t('handbook.englishNotice') }}
      </p>

      <div class="handbook-api-bar">
        <div class="handbook-api-bar-text min-w-0">
          <div class="text-xs font-medium text-[color:var(--muted)]">{{ t('handbook.apiBaseLabel') }}</div>
          <code class="handbook-api-url mt-1 block text-sm font-semibold text-[color:var(--foreground)]">{{ apiBase }}</code>
          <p class="mt-2 text-xs leading-5 text-[color:var(--muted)]">{{ t('handbook.apiBaseHint') }}</p>
        </div>
        <button type="button" class="btn btn-secondary btn-sm shrink-0" @click="copyApiBase">
          {{ apiBaseCopied ? t('handbook.copied') : t('handbook.copy') }}
        </button>
      </div>

      <div
        ref="contentRef"
        class="handbook-content"
        v-html="renderedHtml"
        @click="onContentClick"
      ></div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import DOMPurify from 'dompurify'
import { useAppStore } from '@/stores'
import { resolvePublicApiBaseUrl } from '@/utils/publicApiBaseUrl'
import handbookZh from '@/assets/handbook/handbook.zh.html?raw'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

const contentRef = ref<HTMLElement | null>(null)
const activeHeadingId = ref('')
const apiBaseCopied = ref(false)
let copyTimer = 0

const localeStartsWithEn = computed(() => String(locale.value || '').toLowerCase().startsWith('en'))

const apiBase = computed(() => resolvePublicApiBaseUrl(appStore.apiBaseUrl))
const apiHost = computed(() => {
  const base = apiBase.value
  if (!base) return ''
  try {
    return new URL(base).host
  } catch {
    return base.replace(/^https?:\/\//, '').replace(/\/+$/, '')
  }
})

interface TocItem {
  id: string
  text: string
}

const tocItems = computed<TocItem[]>(() => {
  const source = applyPlaceholders(handbookZh)
  const matches = source.matchAll(/<h2\s+id="([^"]+)"[^>]*>([\s\S]*?)<\/h2>/gi)
  return [...matches].map((match) => ({
    id: match[1],
    text: match[2].replace(/<[^>]+>/g, '').trim(),
  }))
})

function applyPlaceholders(source: string): string {
  return source.replace(/\{\{API_BASE\}\}/g, apiBase.value).replace(/\{\{API_HOST\}\}/g, apiHost.value)
}

const renderedHtml = computed(() => {
  const html = applyPlaceholders(handbookZh)
  return DOMPurify.sanitize(html, {
    ADD_TAGS: ['figure', 'figcaption'],
    ADD_ATTR: ['id', 'class', 'loading', 'download', 'target', 'rel', 'alt'],
  })
})

function scrollToHeading(id: string) {
  const el = document.getElementById(id)
  if (!el) return
  el.scrollIntoView({ behavior: 'smooth', block: 'start' })
  activeHeadingId.value = id
  if (route.hash !== `#${id}`) {
    router.replace({ hash: `#${id}` })
  }
}

function onContentClick(event: MouseEvent) {
  const anchor = (event.target as HTMLElement | null)?.closest('a')
  if (!anchor) return
  const href = anchor.getAttribute('href') || ''
  if (href.startsWith('#') && href.length > 1) {
    event.preventDefault()
    scrollToHeading(href.slice(1))
    return
  }
  if (!href.startsWith('/') || href.startsWith('//') || href.startsWith('/handbook/')) {
    return
  }
  event.preventDefault()
  router.push(href)
}

async function copyApiBase() {
  try {
    await navigator.clipboard.writeText(apiBase.value)
    apiBaseCopied.value = true
    window.clearTimeout(copyTimer)
    copyTimer = window.setTimeout(() => {
      apiBaseCopied.value = false
    }, 2000)
  } catch {
    apiBaseCopied.value = false
  }
}

function injectCopyButtons() {
  const container = contentRef.value
  if (!container) return
  container.querySelectorAll('pre').forEach((pre) => {
    if (pre.querySelector('.copy-btn')) return
    const btn = document.createElement('button')
    btn.type = 'button'
    btn.className = 'copy-btn'
    btn.textContent = t('handbook.copy')
    btn.addEventListener('click', async () => {
      const code = pre.querySelector('code')?.textContent ?? pre.textContent ?? ''
      try {
        await navigator.clipboard.writeText(code)
        btn.textContent = t('handbook.copied')
        window.setTimeout(() => {
          btn.textContent = t('handbook.copy')
        }, 2000)
      } catch {
        btn.textContent = t('handbook.copyFailed')
        window.setTimeout(() => {
          btn.textContent = t('handbook.copy')
        }, 2000)
      }
    })
    pre.appendChild(btn)
  })
}

function syncActiveHeading() {
  if (tocItems.value.length === 0) return
  const offset = 96
  let current = tocItems.value[0].id
  for (const item of tocItems.value) {
    const el = document.getElementById(item.id)
    if (!el) continue
    if (el.getBoundingClientRect().top - offset <= 0) {
      current = item.id
    }
  }
  activeHeadingId.value = current
}

function applyInitialHash() {
  const id = decodeURIComponent((route.hash || '').replace(/^#/, ''))
  if (!id) return
  const el = document.getElementById(id)
  if (el) {
    el.scrollIntoView({ behavior: 'auto', block: 'start' })
    activeHeadingId.value = id
  }
}

onMounted(async () => {
  await appStore.fetchPublicSettings().catch(() => undefined)
  await nextTick()
  injectCopyButtons()
  applyInitialHash()
  syncActiveHeading()
  window.addEventListener('scroll', syncActiveHeading, { passive: true })
})

watch(renderedHtml, async () => {
  await nextTick()
  injectCopyButtons()
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', syncActiveHeading)
  window.clearTimeout(copyTimer)
})
</script>

<style scoped>
.handbook-page {
  display: grid;
  grid-template-columns: minmax(16rem, 18rem) minmax(0, 1fr);
  gap: 28px;
  align-items: start;
  width: 100%;
  max-width: 1200px;
  margin-inline: auto;
}

@media (max-width: 960px) {
  .handbook-page {
    grid-template-columns: 1fr;
  }

  .handbook-toc {
    position: static;
    order: -1;
  }
}

.handbook-toc {
  position: sticky;
  top: 5.5rem;
  width: 100%;
  min-width: 0;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--surface);
  padding: 16px 14px;
}

.handbook-toc-title {
  margin-bottom: 10px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.04em;
  color: var(--muted);
  text-transform: uppercase;
}

.handbook-toc-nav {
  display: flex;
  flex-direction: column;
  gap: 2px;
  max-height: calc(100vh - 10rem);
  overflow: auto;
}

.handbook-toc-link {
  border-radius: var(--radius-sm);
  padding: 6px 8px;
  color: var(--foreground-secondary);
  font-size: 13px;
  line-height: 1.4;
  text-decoration: none;
}

.handbook-toc-link:hover,
.handbook-toc-link.is-active {
  background: var(--surface-deep);
  color: var(--accent-ink);
}

.handbook-main {
  min-width: 0;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--surface);
  padding: 32px 36px 48px;
  box-shadow: var(--shadow-sm, 0 1px 2px rgba(16, 24, 40, 0.04));
}

@media (max-width: 640px) {
  .handbook-main {
    padding: 20px 16px 32px;
  }
}

.handbook-api-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 28px;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--surface-deep);
  padding: 14px 16px;
}

.handbook-api-bar-text {
  flex: 1 1 16rem;
  min-width: 0;
}

.handbook-api-url {
  overflow-wrap: break-word;
  word-break: normal;
}

.handbook-content {
  font-size: 16px;
  line-height: 1.8;
  overflow-wrap: break-word;
  color: var(--foreground-secondary);
}

.handbook-content :deep(h1),
.handbook-content :deep(h2),
.handbook-content :deep(h3),
.handbook-content :deep(h4) {
  color: var(--foreground);
  font-weight: 600;
  letter-spacing: 0;
  scroll-margin-top: 5.5rem;
}

.handbook-content :deep(h1) {
  margin: 0 0 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--border);
  font-size: 32px;
  line-height: 1.25;
}

.handbook-content :deep(h2) {
  margin: 2rem 0 0.75rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--border);
  font-size: 22px;
}

.handbook-content :deep(h3) {
  margin: 1.5rem 0 0.5rem;
  font-size: 1.15rem;
}

.handbook-content :deep(p),
.handbook-content :deep(.lead) {
  margin-bottom: 1rem;
}

.handbook-content :deep(a) {
  color: var(--accent-ink);
  text-decoration: underline;
  text-underline-offset: 4px;
}

.handbook-content :deep(a:hover) {
  color: var(--accent-hover);
}

.handbook-content :deep(ul) {
  margin-bottom: 1rem;
  list-style: disc;
  padding-left: 1.5rem;
}

.handbook-content :deep(ol) {
  margin-bottom: 1rem;
  list-style: decimal;
  padding-left: 1.5rem;
}

.handbook-content :deep(li) {
  margin-bottom: 0.25rem;
}

.handbook-content :deep(code),
.handbook-content :deep(kbd) {
  border-radius: var(--radius-sm);
  background: var(--surface-deep);
  padding: 2px 6px;
  font-family: var(--font-mono), ui-monospace, monospace;
  font-size: 0.875rem;
  color: var(--foreground);
}

.handbook-content :deep(pre) {
  position: relative;
  margin: 1.25rem 0;
  overflow-x: auto;
  border-radius: var(--radius);
  background: var(--surface-deep);
  padding: 16px 18px;
  color: var(--foreground);
  border: 1px solid var(--border);
}

.handbook-content :deep(pre code) {
  background: transparent;
  padding: 0;
  color: inherit;
}

.handbook-content :deep(.copy-btn) {
  position: absolute;
  top: 8px;
  right: 8px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--surface);
  color: var(--foreground-secondary);
  font-size: 12px;
  line-height: 1;
  padding: 5px 8px;
  cursor: pointer;
}

.handbook-content :deep(.copy-btn:hover) {
  color: var(--accent-ink);
}

.handbook-content :deep(table) {
  width: 100%;
  margin: 1.25rem 0;
  border-collapse: collapse;
}

.handbook-content :deep(.table-wrap) {
  width: 100%;
  overflow-x: auto;
}

.handbook-content :deep(th),
.handbook-content :deep(td) {
  border: 1px solid var(--border);
  padding: 8px 12px;
}

.handbook-content :deep(th) {
  background: var(--surface-deep);
  text-align: left;
  font-weight: 600;
  color: var(--foreground);
}

.handbook-content :deep(img) {
  margin: 0;
  height: auto;
  max-width: 100%;
  border-radius: var(--radius);
  border: 1px solid var(--border);
  background: var(--surface);
}

.handbook-content :deep(figure.shot) {
  margin: 1.25rem 0 1.5rem;
}

.handbook-content :deep(figcaption) {
  margin-top: 0.5rem;
  color: var(--muted);
  font-size: 13px;
}

.handbook-content :deep(.notice),
.handbook-content :deep(.callout) {
  margin: 1.25rem 0;
  padding: 12px 16px;
  color: var(--foreground-secondary);
  background: var(--surface-deep);
  border: 1px solid var(--border);
  border-left: 3px solid var(--accent-ink);
  border-radius: var(--radius-sm);
}

.handbook-content :deep(.tips) {
  display: grid;
  gap: 8px;
  margin: 1rem 0 1.25rem;
}

.handbook-content :deep(.tip) {
  display: grid;
  gap: 4px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--surface-deep);
  padding: 10px 12px;
}

.handbook-content :deep(.tip strong) {
  color: var(--foreground);
}

.handbook-content :deep(.model-ids) {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin: 0.75rem 0 1rem;
}

.handbook-content :deep(.model-ids code) {
  font-size: 0.8rem;
}
</style>
