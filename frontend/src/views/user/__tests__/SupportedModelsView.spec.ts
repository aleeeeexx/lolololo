import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import SupportedModelsView from '../SupportedModelsView.vue'
import type { SupportedModelItem } from '@/api/supportedModels'

const { getSupportedModels, copyToClipboard } = vi.hoisted(() => ({
  getSupportedModels: vi.fn(),
  copyToClipboard: vi.fn(),
}))

vi.mock('@/api/supportedModels', () => ({ getSupportedModels }))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copyToClipboard }) }))
vi.mock('@/components/layout/AppLayout.vue', () => ({ default: { template: '<main><slot /></main>' } }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string, params?: Record<string, unknown>) => {
  if (!params) return key
  return `${key}:${JSON.stringify(params)}`
} }) }))

function catalog(items: SupportedModelItem[]): SupportedModelItem[] {
  return items
}

const defaultCatalog: SupportedModelItem[] = [
  { name: 'gpt-4o', platforms: ['openai'], kinds: ['text'], company: 'openai' },
  { name: 'claude-sonnet-4', platforms: ['anthropic'], kinds: ['text'], company: 'anthropic' },
  { name: 'dall-e-3', platforms: ['openai'], kinds: ['image'], company: 'openai' },
  { name: 'sora-2', platforms: ['openai'], kinds: ['video'], company: 'openai' },
  { name: 'grok-imagine-video', platforms: ['grok'], kinds: ['video'], company: 'xai' },
  { name: 'mystery-custom-model', platforms: ['grok'], kinds: ['text'], company: 'other' },
]

const mounted: ReturnType<typeof mount>[] = []
function render() {
  const wrapper = mount(SupportedModelsView, {
    global: { stubs: { Icon: true } },
  })
  mounted.push(wrapper)
  return wrapper
}

beforeEach(() => {
  vi.clearAllMocks()
  getSupportedModels.mockResolvedValue(catalog(defaultCatalog))
  copyToClipboard.mockResolvedValue(true)
})

afterEach(() => {
  mounted.splice(0).forEach((wrapper) => wrapper.unmount())
})

describe('SupportedModelsView', () => {
  it('shows a loading state until models arrive, then lists copyable IDs', async () => {
    let resolveList!: (value: SupportedModelItem[]) => void
    getSupportedModels.mockReturnValueOnce(new Promise<SupportedModelItem[]>((resolve) => { resolveList = resolve }))
    const wrapper = render()
    await flushPromises()
    expect(wrapper.find('[data-testid="supported-models-loading"]').exists()).toBe(true)

    resolveList(catalog(defaultCatalog))
    await flushPromises()
    expect(wrapper.find('[data-testid="supported-models-loading"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="supported-models-source"]').text()).toBe('supportedModels.sourceHint')
    expect(wrapper.get('[data-testid="supported-model-gpt-4o"]').text()).toContain('gpt-4o')
    expect(wrapper.get('[data-testid="supported-model-claude-sonnet-4"]').text()).toContain('claude-sonnet-4')
    expect(wrapper.find('[data-testid="supported-models-unavailable"]').exists()).toBe(false)
  })

  it('still loads models when model plaza and available channels flags would be off', async () => {
    const wrapper = render()
    await flushPromises()
    expect(getSupportedModels).toHaveBeenCalledTimes(1)
    expect(wrapper.get('[data-testid="supported-model-gpt-4o"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="supported-models-unavailable"]').exists()).toBe(false)
  })

  it('filters the list as the user types', async () => {
    const wrapper = render()
    await flushPromises()
    await wrapper.get('[data-testid="supported-models-search"]').setValue('claude')
    expect(wrapper.find('[data-testid="supported-model-gpt-4o"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="supported-model-claude-sonnet-4"]').exists()).toBe(true)
  })

  it('switches by model type and keeps unmatched filters empty', async () => {
    const wrapper = render()
    await flushPromises()
    await wrapper.get('[data-testid="supported-models-kind-image"]').trigger('click')
    expect(wrapper.find('[data-testid="supported-model-dall-e-3"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="supported-model-gpt-4o"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="supported-model-sora-2"]').exists()).toBe(false)

    await wrapper.get('[data-testid="supported-models-kind-video"]').trigger('click')
    expect(wrapper.find('[data-testid="supported-model-sora-2"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="supported-model-grok-imagine-video"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="supported-model-dall-e-3"]').exists()).toBe(false)
  })

  it('switches by company and only lists companies that appear in the catalog', async () => {
    const wrapper = render()
    await flushPromises()
    expect(wrapper.find('[data-testid="supported-models-company-openai"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="supported-models-company-anthropic"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="supported-models-company-xai"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="supported-models-company-other"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="supported-models-company-google"]').exists()).toBe(false)

    await wrapper.get('[data-testid="supported-models-company-xai"]').trigger('click')
    expect(wrapper.find('[data-testid="supported-model-grok-imagine-video"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="supported-model-gpt-4o"]').exists()).toBe(false)
  })

  it('combines type and company filters with AND and shows an empty state instead of the full list', async () => {
    const wrapper = render()
    await flushPromises()
    await wrapper.get('[data-testid="supported-models-kind-video"]').trigger('click')
    await wrapper.get('[data-testid="supported-models-company-openai"]').trigger('click')
    expect(wrapper.find('[data-testid="supported-model-sora-2"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="supported-model-grok-imagine-video"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="supported-models-empty"]').exists()).toBe(false)

    await wrapper.get('[data-testid="supported-models-company-anthropic"]').trigger('click')
    expect(wrapper.get('[data-testid="supported-models-empty"]').text()).toBe('supportedModels.noSearchResult')
    expect(wrapper.find('[data-testid="supported-models-list"]').exists()).toBe(false)
  })

  it('copies the exact model ID when a row is clicked, including after filters are applied', async () => {
    const wrapper = render()
    await flushPromises()
    await wrapper.get('[data-testid="supported-models-kind-image"]').trigger('click')
    await wrapper.get('[data-testid="supported-model-dall-e-3"]').trigger('click')
    expect(copyToClipboard).toHaveBeenCalledWith('dall-e-3', expect.stringContaining('dall-e-3'))
  })

  it('shows an empty state when the API returns no models', async () => {
    getSupportedModels.mockResolvedValueOnce([])
    const wrapper = render()
    await flushPromises()
    expect(wrapper.get('[data-testid="supported-models-empty"]').text()).toBe('supportedModels.empty')
    expect(wrapper.find('[data-testid="supported-models-unavailable"]').exists()).toBe(false)
  })

  it('shows an error state when the API fails, including unauthenticated 401s', async () => {
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    getSupportedModels.mockRejectedValueOnce(Object.assign(new Error('unauthorized'), { status: 401 }))
    const wrapper = render()
    await flushPromises()
    expect(wrapper.get('[data-testid="supported-models-error"]').text()).toBe('supportedModels.loadFailed')
    errorSpy.mockRestore()
  })
})
