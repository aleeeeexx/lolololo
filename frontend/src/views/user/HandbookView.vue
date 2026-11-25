<template>
  <AppLayout v-if="isAuthenticated">
    <HandbookArticle />
  </AppLayout>
  <div v-else class="min-h-screen bg-[color:var(--background)] text-[color:var(--foreground)]">
    <header class="app-topbar">
      <div class="content-wrap flex h-full items-center justify-between gap-4">
        <RouterLink to="/home" class="flex min-w-0 items-center gap-3">
          <span class="brand-mark">
            <img :src="siteLogo || '/logo.svg'" alt="Logo" />
          </span>
          <span class="truncate text-[13px] font-semibold text-[color:var(--foreground)]">
            {{ siteName }}
          </span>
        </RouterLink>
        <div class="flex items-center gap-2">
          <LocaleSwitcher />
          <RouterLink
            to="/login"
            class="inline-flex flex-shrink-0 items-center justify-center rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white shadow-sm shadow-primary-600/20 transition hover:bg-primary-700"
          >
            {{ t('home.login') }}
          </RouterLink>
        </div>
      </div>
    </header>
    <div class="px-5 py-6 md:px-8 md:py-8">
      <HandbookArticle />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { useAuthStore, useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import { DEFAULT_SITE_NAME } from '@/constants/branding'
import HandbookArticle from '@/components/user/HandbookArticle.vue'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const isAuthenticated = computed(() => authStore.isAuthenticated)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || DEFAULT_SITE_NAME)
const siteLogo = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', {
    allowRelative: true,
    allowDataUrl: true,
  }),
)
</script>
