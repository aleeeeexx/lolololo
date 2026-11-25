import { afterEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ checkAuth: vi.fn(), isAuthenticated: false, isAdmin: false, isSimpleMode: false }) }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ siteName: 'DubDub', backendModeEnabled: false, cachedPublicSettings: null }) }))
vi.mock('@/stores/adminSettings', () => ({ useAdminSettingsStore: () => ({ customMenuItems: [] }) }))
vi.mock('@/stores/adminCompliance', () => ({ useAdminComplianceStore: () => ({}) }))
vi.mock('@/composables/useNavigationLoading', () => ({
  useNavigationLoadingState: () => ({ startNavigation: vi.fn(), endNavigation: vi.fn() })
}))
vi.mock('@/composables/useRoutePrefetch', () => ({ useRoutePrefetch: () => ({ triggerPrefetch: vi.fn() }) }))
vi.mock('@/views/auth/LoginView.vue', () => ({ default: { template: '<div>Login</div>' } }))
vi.mock('@/router/title', () => ({ resolveRouteDocumentTitle: () => 'Handbook' }))

afterEach(() => { vi.restoreAllMocks() })

describe('Handbook route', () => {
  it('is a public in-site handbook and is not admin-only', async () => {
    vi.spyOn(window, 'scrollTo').mockImplementation(() => {})
    const { default: router } = await import('@/router')
    const record = router.getRoutes().find(route => route.path === '/handbook')
    expect(record?.name).toBe('Handbook')
    expect(router.getRoutes().some(route => route.path === '/docs/handbook' && route.name === 'Handbook')).toBe(true)
    expect(record?.meta.requiresAuth).toBe(false)
    expect(record?.meta.requiresAdmin).toBeFalsy()
    expect(record?.meta.titleKey).toBe('handbook.title')
    expect(record?.meta.descriptionKey).toBe('handbook.description')
    await router.push('/handbook')
    expect(router.currentRoute.value.path).toBe('/handbook')
  })

  it('resolves the docs alias to the same handbook page', async () => {
    vi.spyOn(window, 'scrollTo').mockImplementation(() => {})
    const { default: router } = await import('@/router')
    await router.push('/docs/handbook')
    expect(router.currentRoute.value.name).toBe('Handbook')
    expect(['/handbook', '/docs/handbook']).toContain(router.currentRoute.value.path)
  })
})
