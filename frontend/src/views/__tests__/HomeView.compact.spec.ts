import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'

import HomeView from '../HomeView.vue'

const { appStore, authStore } = vi.hoisted(() => ({
  appStore: {
    cachedPublicSettings: {} as Record<string, unknown>,
    siteName: 'Fallback site',
    siteLogo: '',
    docUrl: '',
    publicSettingsLoaded: true,
    fetchPublicSettings: vi.fn(),
  },
  authStore: {
    isAuthenticated: false,
    isAdmin: false,
    user: null as { email?: string } | null,
    checkAuth: vi.fn(),
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => appStore,
  useAuthStore: () => authStore,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStore,
}))

const { getPublicGroupRates } = vi.hoisted(() => ({
  getPublicGroupRates: vi.fn(),
}))

vi.mock('@/api/publicGroupRates', () => ({ getPublicGroupRates }))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (!params) return key
        return `${key}:${JSON.stringify(params)}`
      },
    }),
  }
})

function mountHome(settings: Record<string, unknown> = {}) {
  appStore.cachedPublicSettings = {
    site_name: 'Test site',
    site_subtitle: 'Test subtitle',
    ...settings,
  }

  return mount(HomeView, {
    global: {
      stubs: {
        RouterLink: RouterLinkStub,
        LocaleSwitcher: { template: '<div data-testid="locale-switcher" />' },
        Icon: { template: '<span data-testid="icon" />' },
      },
    },
  })
}

function compactDestination(wrapper: ReturnType<typeof mountHome>) {
  return wrapper
    .get('[data-testid="compact-home"]')
    .findAllComponents(RouterLinkStub)
    .find((link) => {
      const to = link.props('to')
      return to === '/login' || to === '/dashboard' || to === '/admin/dashboard'
    })
    ?.props('to')
}

function handbookDestination(wrapper: ReturnType<typeof mountHome>) {
  return wrapper
    .findAllComponents(RouterLinkStub)
    .find((link) => link.props('to') === '/handbook')
    ?.props('to')
}

function modelPlazaDestination(wrapper: ReturnType<typeof mountHome>) {
  return wrapper
    .findAllComponents(RouterLinkStub)
    .find((link) => link.props('to') === '/model-plaza')
    ?.props('to')
}

describe('HomeView compact mode', () => {
  beforeEach(() => {
    authStore.isAuthenticated = false
    authStore.isAdmin = false
    authStore.user = null
    authStore.checkAuth.mockClear()
    appStore.fetchPublicSettings.mockClear()
    getPublicGroupRates.mockReset()
    getPublicGroupRates.mockResolvedValue([
      {
        id: 1,
        name: 'default',
        rate_multiplier: 1,
        image_rate_independent: false,
        image_rate_multiplier: 1,
        video_rate_independent: false,
        video_rate_multiplier: 1,
      },
      {
        id: 2,
        name: 'promo',
        rate_multiplier: 0.5,
        image_rate_independent: false,
        image_rate_multiplier: 1,
        video_rate_independent: false,
        video_rate_multiplier: 1,
      },
    ])
    localStorage.clear()
    vi.spyOn(window, 'matchMedia').mockReturnValue({ matches: false } as MediaQueryList)
  })

  it('renders custom HTML ahead of compact mode', () => {
    const wrapper = mountHome({
      compact_home_enabled: true,
      home_content: '<section id="custom-home">Custom home</section>',
    })

    expect(wrapper.get('#custom-home').text()).toBe('Custom home')
    expect(wrapper.find('[data-testid="compact-home"]').exists()).toBe(false)
  })

  it('renders custom URL content ahead of compact mode', () => {
    const wrapper = mountHome({
      compact_home_enabled: true,
      home_content: ' https://example.com/home ',
    })

    expect(wrapper.get('iframe').attributes('src')).toBe('https://example.com/home')
    expect(wrapper.find('[data-testid="compact-home"]').exists()).toBe(false)
  })

  it('treats whitespace-only custom content as empty and selects compact mode', () => {
    const wrapper = mountHome({ compact_home_enabled: true, home_content: ' \n\t ' })

    expect(wrapper.get('[data-testid="compact-home"]').text()).toContain('Test site')
  })

  it.each([undefined, false])('selects the default home when compact mode is %s', (enabled) => {
    const settings = enabled === undefined ? {} : { compact_home_enabled: enabled }
    const wrapper = mountHome(settings)

    expect(wrapper.find('[data-testid="compact-home"]').exists()).toBe(false)
    expect(wrapper.find('.terminal-container').exists()).toBe(true)
  })

  it('links unauthenticated visitors to login', () => {
    expect(compactDestination(mountHome({ compact_home_enabled: true }))).toBe('/login')
  })

  it('links authenticated users to their dashboard', () => {
    authStore.isAuthenticated = true

    expect(compactDestination(mountHome({ compact_home_enabled: true }))).toBe('/dashboard')
  })

  it('links administrators to the admin dashboard', () => {
    authStore.isAuthenticated = true
    authStore.isAdmin = true

    const wrapper = mountHome({ compact_home_enabled: true })
    expect(compactDestination(wrapper)).toBe('/admin/dashboard')
    expect(authStore.checkAuth).toHaveBeenCalledOnce()
    expect(appStore.fetchPublicSettings).not.toHaveBeenCalled()
  })

  it('shows the model plaza link to anonymous visitors when public access is enabled', () => {
    const wrapper = mountHome({
      compact_home_enabled: true,
      model_plaza_enabled: true,
      model_plaza_require_auth: false,
    })

    expect(modelPlazaDestination(wrapper)).toBe('/model-plaza')
  })

  it('hides the model plaza link from anonymous visitors when sign-in is required', () => {
    const wrapper = mountHome({
      compact_home_enabled: true,
      model_plaza_enabled: true,
      model_plaza_require_auth: true,
    })

    expect(modelPlazaDestination(wrapper)).toBeUndefined()
  })

  it('shows the model plaza link to authenticated visitors when sign-in is required', () => {
    authStore.isAuthenticated = true

    const wrapper = mountHome({
      compact_home_enabled: true,
      model_plaza_enabled: true,
      model_plaza_require_auth: true,
    })

    expect(modelPlazaDestination(wrapper)).toBe('/model-plaza')
  })

  it('shows the model plaza link in the default home header', () => {
    const wrapper = mountHome({
      model_plaza_enabled: true,
      model_plaza_require_auth: false,
    })

    expect(modelPlazaDestination(wrapper)).toBe('/model-plaza')
  })

  it('always shows an in-site handbook link in compact and default homes', () => {
    expect(handbookDestination(mountHome({ compact_home_enabled: true }))).toBe('/handbook')
    expect(handbookDestination(mountHome())).toBe('/handbook')
  })

  it('hides the model plaza link when the feature is disabled', () => {
    const wrapper = mountHome({
      compact_home_enabled: true,
      model_plaza_enabled: false,
      model_plaza_require_auth: false,
    })

    expect(modelPlazaDestination(wrapper)).toBeUndefined()
  })

  it('shows official savings from the cheapest public group on the compact homepage', async () => {
    const wrapper = mountHome({ compact_home_enabled: true })
    await flushPromises()
    expect(wrapper.get('[data-testid="official-savings-title"]').text()).toContain('officialSavings.cheaperMinimum')
    expect(wrapper.get('[data-testid="official-savings-title"]').text()).toContain('"percent":93')
  })

  it('shows official savings from the cheapest public group on the default homepage', async () => {
    const wrapper = mountHome()
    await flushPromises()
    expect(wrapper.find('[data-testid="compact-home"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="official-savings-title"]').text()).toContain('officialSavings.cheaperMinimum')
    expect(wrapper.get('[data-testid="official-savings-title"]').text()).toContain('"percent":93')
  })

  it('omits the savings claim when public group rates cannot be loaded', async () => {
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    getPublicGroupRates.mockRejectedValueOnce(new Error('network'))
    const wrapper = mountHome()
    await flushPromises()
    expect(wrapper.find('[data-testid="official-savings"]').exists()).toBe(false)
    errorSpy.mockRestore()
  })
})
