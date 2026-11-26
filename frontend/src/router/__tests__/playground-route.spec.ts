import { afterEach, describe, expect, it, vi } from 'vitest'

const authStore = vi.hoisted(() => ({
  checkAuth: vi.fn(),
  isAuthenticated: false,
  isAdmin: false,
  isSimpleMode: false,
  hasPendingAuthSession: false,
}))

vi.mock('@/stores/auth', () => ({ useAuthStore: () => authStore }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ siteName: 'Sub2API', backendModeEnabled: false, cachedPublicSettings: null }) }))
vi.mock('@/stores/adminSettings', () => ({ useAdminSettingsStore: () => ({ customMenuItems: [] }) }))
vi.mock('@/stores/adminCompliance', () => ({
  useAdminComplianceStore: () => ({
    initialized: true,
    fetchStatus: vi.fn(),
    requireAcknowledgement: vi.fn(),
  }),
}))
vi.mock('@/composables/useNavigationLoading', () => ({
  useNavigationLoadingState: () => ({ startNavigation: vi.fn(), endNavigation: vi.fn() })
}))
vi.mock('@/composables/useRoutePrefetch', () => ({ useRoutePrefetch: () => ({ triggerPrefetch: vi.fn() }) }))
vi.mock('@/views/auth/LoginView.vue', () => ({ default: { template: '<div>Login</div>' } }))
vi.mock('@/views/user/DashboardView.vue', () => ({ default: { template: '<div>Dashboard</div>' } }))
vi.mock('@/views/user/UseView.vue', () => ({ default: { template: '<div>Playground</div>' } }))
vi.mock('@/router/title', () => ({ resolveRouteDocumentTitle: () => 'Direct Use' }))

afterEach(() => {
  authStore.isAuthenticated = false
  authStore.isAdmin = false
  vi.restoreAllMocks()
})

describe('Direct Use route', () => {
  it('requires login and is admin-only', async () => {
    vi.spyOn(window, 'scrollTo').mockImplementation(() => {})
    const { default: router } = await import('@/router')
    const record = router.getRoutes().find(route => route.name === 'Playground')
    expect(record?.path).toBe('/use')
    expect(record?.meta.requiresAuth).toBe(true)
    expect(record?.meta.requiresAdmin).toBe(true)
    expect(record?.meta.titleKey).toBe('playground.title')
    await router.push('/use')
    expect(router.currentRoute.value.path).toBe('/login')
    expect(router.currentRoute.value.query.redirect).toBe('/use')
  })

  it('sends a logged-in regular user to the dashboard', async () => {
    vi.spyOn(window, 'scrollTo').mockImplementation(() => {})
    authStore.isAuthenticated = true
    authStore.isAdmin = false
    const { default: router } = await import('@/router')
    await router.push('/use')
    expect(router.currentRoute.value.path).toBe('/dashboard')
  })

  it('lets an administrator open Direct Use', async () => {
    vi.spyOn(window, 'scrollTo').mockImplementation(() => {})
    authStore.isAuthenticated = true
    authStore.isAdmin = true
    const { default: router } = await import('@/router')
    await router.push('/use')
    expect(router.currentRoute.value.path).toBe('/use')
    expect(router.currentRoute.value.name).toBe('Playground')
  })
})
