import { afterEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ checkAuth: vi.fn(), isAuthenticated: false, isAdmin: false, isSimpleMode: false }) }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ siteName: 'Sub2API', backendModeEnabled: false, cachedPublicSettings: null }) }))
vi.mock('@/stores/adminSettings', () => ({ useAdminSettingsStore: () => ({ customMenuItems: [] }) }))
vi.mock('@/stores/adminCompliance', () => ({ useAdminComplianceStore: () => ({}) }))
vi.mock('@/composables/useNavigationLoading', () => ({
  useNavigationLoadingState: () => ({ startNavigation: vi.fn(), endNavigation: vi.fn() })
}))
vi.mock('@/composables/useRoutePrefetch', () => ({ useRoutePrefetch: () => ({ triggerPrefetch: vi.fn() }) }))
vi.mock('@/views/auth/LoginView.vue', () => ({ default: { template: '<div>Login</div>' } }))
vi.mock('@/router/title', () => ({ resolveRouteDocumentTitle: () => 'Supported Models' }))

afterEach(() => { vi.restoreAllMocks() })

describe('Supported Models route', () => {
  it('is a logged-in user route and is not admin-only', async () => {
    vi.spyOn(window, 'scrollTo').mockImplementation(() => {})
    const { default: router } = await import('@/router')
    const record = router.getRoutes().find(route => route.name === 'SupportedModels')
    expect(record?.path).toBe('/supported-models')
    expect(record?.meta.requiresAuth).toBe(true)
    expect(record?.meta.requiresAdmin).toBe(false)
    expect(record?.meta.titleKey).toBe('supportedModels.title')
    await router.push('/supported-models')
    expect(router.currentRoute.value.path).toBe('/login')
    expect(router.currentRoute.value.query.redirect).toBe('/supported-models')
  })

  it('keeps the existing model plaza route as a public entry', async () => {
    const { default: router } = await import('@/router')
    const plaza = router.getRoutes().find(route => route.name === 'ModelPlaza')
    expect(plaza?.path).toBe('/model-plaza')
    expect(plaza?.meta.requiresAuth).toBe(false)
  })
})
