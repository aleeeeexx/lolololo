<template>
  <header class="app-topbar">
    <div class="content-wrap flex h-full items-center justify-between gap-4">
      <!-- 左:站点 logo + 名称 -->
      <div class="flex min-w-0 items-center gap-3">
        <template v-if="settings">
          <span class="brand-mark">
            <img :src="siteLogo || '/logo.svg'" alt="Logo" />
          </span>
          <span class="truncate text-[13px] font-semibold text-[color:var(--foreground)]">
            {{ siteName }}
          </span>
        </template>
        <template v-else>
          <span class="brand-mark animate-pulse" aria-hidden="true"></span>
          <span class="h-5 w-28 animate-pulse rounded bg-[color:var(--surface-deep)]" aria-hidden="true"></span>
        </template>
      </div>

      <!-- 右:登录 / 回到后台 -->
      <RouterLink
        v-if="isAuthenticated"
        :to="backTarget"
        class="btn btn-primary flex-shrink-0"
      >
        {{ t('modelPlaza.nav.backToDashboard') }}
      </RouterLink>
      <RouterLink
        v-else
        :to="{ path: '/login', query: { redirect: '/model-plaza' } }"
        class="btn btn-primary flex-shrink-0"
      >
        {{ t('modelPlaza.nav.login') }}
      </RouterLink>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { sanitizeUrl } from '@/utils/url'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { DEFAULT_SITE_NAME } from '@/constants/branding'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const settings = computed(() => appStore.cachedPublicSettings)
const siteName = computed(() => settings.value?.site_name || DEFAULT_SITE_NAME)
const siteLogo = computed(() =>
  sanitizeUrl(settings.value?.site_logo || '', { allowRelative: true, allowDataUrl: true })
)
const isAuthenticated = computed(() => authStore.isAuthenticated)
const backTarget = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
</script>
