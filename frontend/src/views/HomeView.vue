<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="hasHomeContent" class="min-h-screen">
    <!-- iframe mode -->
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Compact Home Page -->
  <div
    v-else-if="compactHomeEnabled"
    data-testid="compact-home"
    class="flex min-h-screen flex-col bg-[color:var(--background)] text-[color:var(--foreground)]"
  >
    <header class="app-topbar px-5 md:px-8">
      <nav class="content-wrap flex h-full flex-wrap items-center justify-between gap-3 !px-0">
        <div class="flex min-w-0 flex-1 items-center gap-3">
          <span class="brand-mark">
            <img :src="siteLogo || '/logo.svg'" alt="Logo" />
          </span>
          <span class="min-w-0 truncate text-[13px] font-semibold text-[color:var(--foreground)]">{{ siteName }}</span>
        </div>
        <div class="flex max-w-full shrink-0 flex-wrap items-center justify-end gap-2">
          <LocaleSwitcher />
          <router-link
            to="/handbook"
            class="flex h-10 shrink-0 items-center gap-1.5 rounded-[length:var(--radius-sm)] px-2.5 text-[13px] font-medium text-[color:var(--muted)] transition-colors duration-150 hover:bg-[color:var(--surface-deep)] hover:text-[color:var(--foreground)]"
            :title="t('nav.handbook')"
            :aria-label="t('nav.handbook')"
          >
            <Icon name="book" size="md" />
            <span class="hidden sm:inline">{{ t('nav.handbook') }}</span>
          </router-link>
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-[length:var(--radius-sm)] text-[13px] text-[color:var(--muted)] transition-colors duration-150 hover:bg-[color:var(--surface-deep)] hover:text-[color:var(--foreground)]"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>
          <router-link
            v-if="showModelPlazaEntry"
            to="/model-plaza"
            class="flex h-10 shrink-0 items-center gap-1.5 rounded-[length:var(--radius-sm)] px-2.5 text-[13px] font-medium text-[color:var(--muted)] transition-colors duration-150 hover:bg-[color:var(--surface-deep)] hover:text-[color:var(--foreground)]"
            :title="t('nav.modelPlaza')"
          >
            <Icon name="grid" size="md" />
            <span class="hidden sm:inline">{{ t('nav.modelPlaza') }}</span>
          </router-link>
          <button
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-[length:var(--radius-sm)] text-[color:var(--muted)] transition-colors duration-150 hover:bg-[color:var(--surface-deep)] hover:text-[color:var(--foreground)]"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="toggleTheme"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="btn btn-primary shrink-0"
          >
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="flex min-w-0 flex-1 items-center justify-center px-5 py-16 md:px-8">
      <div class="min-w-0 max-w-2xl text-center">
        <span class="brand-mark mx-auto mb-6 !h-12 !w-12">
          <img :src="siteLogo || '/logo.svg'" alt="Logo" />
        </span>
        <h1 class="hero-title [overflow-wrap:anywhere]">{{ siteName }}</h1>
        <p class="mt-4 whitespace-pre-wrap [overflow-wrap:anywhere] text-base text-[color:var(--foreground-secondary)]">{{ siteSubtitle }}</p>
        <OfficialSavingsCallout
          v-if="publicRateGroups.length > 0"
          class="mt-5 text-left sm:text-center"
          variant="hero"
          :groups="publicRateGroups"
        />
        <router-link
          :to="isAuthenticated ? dashboardPath : '/login'"
          class="btn btn-primary mt-8"
        >
          {{ isAuthenticated ? t('home.goToDashboard') : t('home.login') }}
        </router-link>
      </div>
    </main>

    <footer class="site-footer min-w-0 px-5 py-5 text-center text-[13px] text-[color:var(--muted)] [overflow-wrap:anywhere] md:px-8">
      &copy; {{ currentYear }} {{ siteName }}
    </footer>
  </div>

  <!-- Default Home Page -->
  <div
    v-else
    class="relative flex min-h-screen flex-col bg-[color:var(--background)]"
  >
    <!-- Header -->
    <header class="app-topbar px-5 md:px-8">
      <nav class="content-wrap flex h-full items-center justify-between !px-0">
        <!-- Logo -->
        <div class="flex items-center gap-3">
          <span class="brand-mark">
            <img :src="siteLogo || '/logo.svg'" alt="Logo" />
          </span>
          <span class="hidden text-[13px] font-semibold text-[color:var(--foreground)] sm:inline">{{ siteName }}</span>
        </div>

        <!-- Nav Actions -->
        <div class="flex items-center gap-3">
          <!-- Language Switcher -->
          <LocaleSwitcher />

          <!-- In-site handbook -->
          <router-link
            to="/handbook"
            class="inline-flex items-center gap-1.5 rounded-[length:var(--radius-sm)] p-2 text-[13px] font-medium text-[color:var(--muted)] transition-colors duration-150 hover:bg-[color:var(--surface-deep)] hover:text-[color:var(--foreground)]"
            :title="t('nav.handbook')"
            :aria-label="t('nav.handbook')"
          >
            <Icon name="book" size="md" />
            <span class="hidden sm:inline">{{ t('nav.handbook') }}</span>
          </router-link>

          <!-- Doc Link -->
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="rounded-[length:var(--radius-sm)] p-2 text-[13px] text-[color:var(--muted)] transition-colors duration-150 hover:bg-[color:var(--surface-deep)] hover:text-[color:var(--foreground)]"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>

          <!-- Model Plaza Link -->
          <router-link
            v-if="showModelPlazaEntry"
            to="/model-plaza"
            class="inline-flex items-center gap-1.5 rounded-[length:var(--radius-sm)] p-2 text-[13px] text-[color:var(--muted)] transition-colors duration-150 hover:bg-[color:var(--surface-deep)] hover:text-[color:var(--foreground)]"
            :title="t('nav.modelPlaza')"
          >
            <Icon name="grid" size="md" />
            <span class="hidden sm:inline">{{ t('nav.modelPlaza') }}</span>
          </router-link>

          <!-- Theme Toggle -->
          <button
            @click="toggleTheme"
            class="rounded-[length:var(--radius-sm)] p-2 text-[13px] text-[color:var(--muted)] transition-colors duration-150 hover:bg-[color:var(--surface-deep)] hover:text-[color:var(--foreground)]"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <!-- Login / Dashboard Button -->
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="btn btn-primary gap-1.5 !h-8 !px-2 !py-0 text-[13px]"
          >
            <span
              class="flex h-5 w-5 items-center justify-center rounded-[length:var(--radius-sm)] bg-[color:var(--primary-foreground)] text-[10px] font-extrabold text-[color:var(--primary)]"
            >
              {{ userInitial }}
            </span>
            <span class="text-[13px] font-semibold">{{ t('home.dashboard') }}</span>
            <svg
              class="h-3 w-3 text-gray-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M4.5 19.5l15-15m0 0H8.25m11.25 0v11.25"
              />
            </svg>
          </router-link>
          <router-link
            v-else
            to="/login"
            class="btn btn-primary !h-8 !px-3 text-[13px]"
          >
            {{ t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <!-- Main Content -->
    <main class="relative z-10 flex-1">
      <div class="content-wrap border-b border-[color:var(--border)] py-16">
        <!-- Hero Section - Left/Right Layout -->
        <div class="flex flex-col items-center justify-between gap-12 lg:flex-row lg:gap-16">
          <!-- Left: Text Content -->
          <div class="flex-1 text-center lg:text-left">
            <h1 class="hero-title mb-4">
              {{ siteName }}
            </h1>
            <p
              class="text-base text-[color:var(--foreground-secondary)] md:text-lg"
              :class="publicRateGroups.length > 0 ? 'mb-4' : 'mb-8'"
            >
              {{ siteSubtitle }}
            </p>
            <OfficialSavingsCallout
              v-if="publicRateGroups.length > 0"
              class="mb-8 lg:text-left"
              variant="hero"
              :groups="publicRateGroups"
            />

            <!-- CTA Button -->
            <div>
              <router-link
                :to="isAuthenticated ? dashboardPath : '/login'"
                class="btn btn-primary"
              >
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
                <Icon name="arrowRight" size="md" class="ml-2" :stroke-width="2" />
              </router-link>
            </div>
          </div>

          <!-- Right: Terminal Animation -->
          <div class="flex flex-1 justify-center lg:justify-end">
            <div class="terminal-container">
              <div class="terminal-window">
                <!-- Window header -->
                <div class="terminal-header">
                  <div class="terminal-buttons">
                    <span class="btn-close"></span>
                    <span class="btn-minimize"></span>
                    <span class="btn-maximize"></span>
                  </div>
                  <span class="terminal-title">terminal</span>
                </div>
                <!-- Terminal content -->
                <div class="terminal-body">
                  <div class="code-line line-1">
                    <span class="code-prompt">$</span>
                    <span class="code-cmd">curl</span>
                    <span class="code-flag">-X POST</span>
                    <span class="code-url">/v1/messages</span>
                  </div>
                  <div class="code-line line-2">
                    <span class="code-comment"># Routing to upstream...</span>
                  </div>
                  <div class="code-line line-3">
                    <span class="code-success">200 OK</span>
                    <span class="code-response">{ "content": "Hello!" }</span>
                  </div>
                  <div class="code-line line-4">
                    <span class="code-prompt">$</span>
                    <span class="cursor"></span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="content-wrap py-16">
        <!-- Feature Tags - Centered -->
        <div class="mb-12 flex flex-wrap items-center justify-center gap-3 md:gap-4">
          <div
            class="inline-flex items-center gap-2 rounded-[2px] border border-[color:var(--border)] bg-[color:var(--surface)] px-3 py-1.5"
          >
            <Icon name="swap" size="sm" class="text-[color:var(--accent-ink)]" />
            <span class="text-[13px] font-medium text-[color:var(--foreground-secondary)]">{{
              t('home.tags.subscriptionToApi')
            }}</span>
          </div>
          <div
            class="inline-flex items-center gap-2 rounded-[2px] border border-[color:var(--border)] bg-[color:var(--surface)] px-3 py-1.5"
          >
            <Icon name="shield" size="sm" class="text-[color:var(--accent-ink)]" />
            <span class="text-[13px] font-medium text-[color:var(--foreground-secondary)]">{{
              t('home.tags.stickySession')
            }}</span>
          </div>
          <div
            class="inline-flex items-center gap-2 rounded-[2px] border border-[color:var(--border)] bg-[color:var(--surface)] px-3 py-1.5"
          >
            <Icon name="chart" size="sm" class="text-[color:var(--accent-ink)]" />
            <span class="text-[13px] font-medium text-[color:var(--foreground-secondary)]">{{
              t('home.tags.realtimeBilling')
            }}</span>
          </div>
        </div>

        <!-- Features Grid -->
        <div class="mb-12 grid gap-3 md:grid-cols-2 lg:grid-cols-3">
          <!-- Feature 1: Unified Gateway -->
          <div
            class="group rounded-[length:var(--radius)] border border-[color:var(--border)] bg-[color:var(--surface)] p-[22px] transition-[border-color] duration-150 hover:border-[color:var(--border-strong)]"
          >
            <div
              class="icon-well mb-4"
            >
              <Icon name="server" size="lg" class="text-[color:var(--accent-ink)]" />
            </div>
            <h3 class="card-title mb-2">
              {{ t('home.features.unifiedGateway') }}
            </h3>
            <p class="text-[13px] leading-relaxed text-[color:var(--foreground-secondary)]">
              {{ t('home.features.unifiedGatewayDesc') }}
            </p>
          </div>

          <!-- Feature 2: Account Pool -->
          <div
            class="group rounded-[length:var(--radius)] border border-[color:var(--border)] bg-[color:var(--surface)] p-[22px] transition-[border-color] duration-150 hover:border-[color:var(--border-strong)]"
          >
            <div
              class="icon-well mb-4"
            >
              <svg
                class="h-6 w-6 text-[color:var(--accent-ink)]"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M18 18.72a9.094 9.094 0 003.741-.479 3 3 0 00-4.682-2.72m.94 3.198l.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0112 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 016 18.719m12 0a5.971 5.971 0 00-.941-3.197m0 0A5.995 5.995 0 0012 12.75a5.995 5.995 0 00-5.058 2.772m0 0a3 3 0 00-4.681 2.72 8.986 8.986 0 003.74.477m.94-3.197a5.971 5.971 0 00-.94 3.197M15 6.75a3 3 0 11-6 0 3 3 0 016 0zm6 3a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0zm-13.5 0a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z"
                />
              </svg>
            </div>
            <h3 class="card-title mb-2">
              {{ t('home.features.multiAccount') }}
            </h3>
            <p class="text-[13px] leading-relaxed text-[color:var(--foreground-secondary)]">
              {{ t('home.features.multiAccountDesc') }}
            </p>
          </div>

          <!-- Feature 3: Billing & Quota -->
          <div
            class="group rounded-[length:var(--radius)] border border-[color:var(--border)] bg-[color:var(--surface)] p-[22px] transition-[border-color] duration-150 hover:border-[color:var(--border-strong)]"
          >
            <div
              class="icon-well mb-4"
            >
              <svg
                class="h-6 w-6 text-[color:var(--accent-ink)]"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z"
                />
              </svg>
            </div>
            <h3 class="card-title mb-2">
              {{ t('home.features.balanceQuota') }}
            </h3>
            <p class="text-[13px] leading-relaxed text-[color:var(--foreground-secondary)]">
              {{ t('home.features.balanceQuotaDesc') }}
            </p>
          </div>
        </div>

        <!-- Supported Providers -->
        <div class="mb-8 text-center">
          <h2 class="page-section-title mb-3">
            {{ t('home.providers.title') }}
          </h2>
          <p class="text-[14px] text-[color:var(--foreground-secondary)]">
            {{ t('home.providers.description') }}
          </p>
        </div>

        <div class="mb-16 flex flex-wrap items-center justify-center gap-4">
          <!-- Claude - Supported -->
          <div
            class="flex items-center gap-2 rounded-[length:var(--radius)] border border-[color:var(--border)] bg-[color:var(--surface)] px-5 py-3"
          >
            <div
              class="icon-well !h-8 !w-8 !text-[11px] font-extrabold"
            >
              <span class="text-[11px] font-extrabold text-[color:var(--accent-ink)]">C</span>
            </div>
            <span class="text-[13px] font-medium text-[color:var(--foreground-secondary)]">{{ t('home.providers.claude') }}</span>
            <span
              class="tag-selected px-1.5 py-0.5 text-[10px] font-medium"
              >{{ t('home.providers.supported') }}</span
            >
          </div>
          <!-- GPT - Supported -->
          <div
            class="flex items-center gap-2 rounded-[length:var(--radius)] border border-[color:var(--border)] bg-[color:var(--surface)] px-5 py-3"
          >
            <div
              class="icon-well !h-8 !w-8 !text-[11px] font-extrabold"
            >
              <span class="text-[11px] font-extrabold text-[color:var(--accent-ink)]">G</span>
            </div>
            <span class="text-[13px] font-medium text-[color:var(--foreground-secondary)]">GPT</span>
            <span
              class="tag-selected px-1.5 py-0.5 text-[10px] font-medium"
              >{{ t('home.providers.supported') }}</span
            >
          </div>
          <!-- Gemini - Supported -->
          <div
            class="flex items-center gap-2 rounded-[length:var(--radius)] border border-[color:var(--border)] bg-[color:var(--surface)] px-5 py-3"
          >
            <div
              class="icon-well !h-8 !w-8 !text-[11px] font-extrabold"
            >
              <span class="text-[11px] font-extrabold text-[color:var(--accent-ink)]">G</span>
            </div>
            <span class="text-[13px] font-medium text-[color:var(--foreground-secondary)]">{{ t('home.providers.gemini') }}</span>
            <span
              class="tag-selected px-1.5 py-0.5 text-[10px] font-medium"
              >{{ t('home.providers.supported') }}</span
            >
          </div>
          <!-- Antigravity - Supported -->
          <div
            class="flex items-center gap-2 rounded-[length:var(--radius)] border border-[color:var(--border)] bg-[color:var(--surface)] px-5 py-3"
          >
            <div
              class="icon-well !h-8 !w-8 !text-[11px] font-extrabold"
            >
              <span class="text-[11px] font-extrabold text-[color:var(--accent-ink)]">A</span>
            </div>
            <span class="text-[13px] font-medium text-[color:var(--foreground-secondary)]">{{ t('home.providers.antigravity') }}</span>
            <span
              class="tag-selected px-1.5 py-0.5 text-[10px] font-medium"
              >{{ t('home.providers.supported') }}</span
            >
          </div>
          <!-- More - Coming Soon -->
          <div
            class="flex items-center gap-2 rounded-[length:var(--radius)] border border-[color:var(--border)] bg-[color:var(--surface)] px-5 py-3 opacity-60"
          >
            <div
              class="icon-well !h-8 !w-8 !text-[11px] font-extrabold"
            >
              <span class="text-[11px] font-extrabold text-[color:var(--accent-ink)]">+</span>
            </div>
            <span class="text-[13px] font-medium text-[color:var(--foreground-secondary)]">{{ t('home.providers.more') }}</span>
            <span
              class="rounded bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-500 dark:bg-dark-700 dark:text-dark-400"
              >{{ t('home.providers.soon') }}</span
            >
          </div>
        </div>
      </div>
    </main>

    <!-- Footer -->
    <footer class="site-footer relative z-10 px-5 py-8 md:px-8">
      <div
        class="content-wrap flex flex-col items-center justify-center gap-4 text-center !px-0 sm:flex-row sm:text-left"
      >
        <p class="text-[13px] text-[color:var(--muted)]">
          &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
        </p>
        <div class="flex items-center gap-4">
          <router-link
            to="/handbook"
            class="text-[13px] text-[color:var(--accent-ink)] transition-colors duration-150 hover:text-[color:var(--accent-hover)]"
          >
            {{ t('nav.handbook') }}
          </router-link>
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-[13px] text-[color:var(--accent-ink)] transition-colors duration-150 hover:text-[color:var(--accent-hover)]"
          >
            {{ t('home.docs') }}
          </a>
          <a
            :href="githubUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-[13px] text-[color:var(--accent-ink)] transition-colors duration-150 hover:text-[color:var(--accent-hover)]"
          >
            GitHub
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import OfficialSavingsCallout from '@/components/common/OfficialSavingsCallout.vue'
import Icon from '@/components/icons/Icon.vue'
import { getPublicGroupRates, type PublicGroupRate } from '@/api/publicGroupRates'
import { sanitizeUrl } from '@/utils/url'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import { DEFAULT_SITE_NAME, DEFAULT_SITE_SUBTITLE } from '@/constants/branding'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || DEFAULT_SITE_NAME)
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || DEFAULT_SITE_SUBTITLE)
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const hasHomeContent = computed(() => homeContent.value.trim().length > 0)
const compactHomeEnabled = computed(() => appStore.cachedPublicSettings?.compact_home_enabled === true)
const modelPlazaEnabled = computed(() => isFeatureFlagEnabled(FeatureFlags.modelPlaza))

// Check if homeContent is a URL (for iframe display)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))

// GitHub URL
const githubUrl = 'https://github.com/Wei-Shaw/sub2api'

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const modelPlazaRequiresAuth = computed(
  () => appStore.cachedPublicSettings?.model_plaza_require_auth === true,
)
const showModelPlazaEntry = computed(
  () => modelPlazaEnabled.value && (isAuthenticated.value || !modelPlazaRequiresAuth.value),
)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

// Current year for footer
const currentYear = computed(() => new Date().getFullYear())

// Toggle theme
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Initialize theme
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme === 'dark') {
    isDark.value = true
    document.documentElement.classList.add('dark')
  } else {
    isDark.value = false
    document.documentElement.classList.remove('dark')
  }
}

const publicRateGroups = ref<PublicGroupRate[]>([])

async function loadPublicGroupRates() {
  try {
    publicRateGroups.value = await getPublicGroupRates()
  } catch (err: unknown) {
    console.error('Failed to load public group rates:', err)
    publicRateGroups.value = []
  }
}

onMounted(() => {
  initTheme()

  // Check auth state
  authStore.checkAuth()

  // Ensure public settings are loaded (will use cache if already loaded from injected config)
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }

  void loadPublicGroupRates()
})
</script>

<style scoped>
.terminal-container {
  position: relative;
  display: inline-block;
}

.terminal-window {
  width: 420px;
  max-width: 100%;
  background: #0c0c0c;
  border-radius: var(--radius);
  border: 1px solid var(--border);
  box-shadow: none;
  overflow: hidden;
}

.terminal-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: #161618;
  border-bottom: 1px solid #252628;
}

.terminal-buttons {
  display: flex;
  gap: 8px;
}

.terminal-buttons span {
  width: 10px;
  height: 10px;
  border-radius: 2px;
  background: #9b9ba6;
}

.btn-close,
.btn-minimize,
.btn-maximize {
  background: #9b9ba6;
}

.terminal-title {
  flex: 1;
  text-align: center;
  font-size: 12px;
  font-family: var(--font-mono), ui-monospace, monospace;
  color: #9b9ba6;
  margin-right: 52px;
}

.terminal-body {
  padding: 20px 24px;
  font-family: var(--font-mono), ui-monospace, monospace;
  font-size: 14px;
  line-height: 2;
  color: #cccccc;
}

.code-line {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.code-prompt,
.code-success {
  color: #f0f0fa;
  font-weight: 600;
}

.code-cmd,
.code-url,
.code-response {
  color: #cccccc;
}

.code-flag,
.code-comment {
  color: #9b9ba6;
}

.code-comment {
  font-style: italic;
}

.code-success {
  background: rgba(240, 240, 250, 0.08);
  padding: 2px 8px;
  border-radius: 2px;
}

.cursor {
  display: inline-block;
  width: 8px;
  height: 16px;
  background: #cccccc;
}

@media (prefers-reduced-motion: no-preference) {
  .code-line {
    opacity: 0;
    animation: line-appear 0.15s ease forwards;
  }

  .line-1 { animation-delay: 0.15s; }
  .line-2 { animation-delay: 0.3s; }
  .line-3 { animation-delay: 0.45s; }
  .line-4 { animation-delay: 0.6s; }

  .cursor {
    animation: blink 1s step-end infinite;
  }
}

@keyframes line-appear {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}
</style>
