import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import UseView from '../UseView.vue'
import type { PlaygroundSession } from '@/api/playground'

const { getPlaygroundSession } = vi.hoisted(() => ({
  getPlaygroundSession: vi.fn(),
}))

vi.mock('@/api/playground', () => ({
  getPlaygroundSession,
  fetchPlaygroundImageContent: vi.fn(),
  getPlaygroundImageTask: vi.fn(),
  pollPlaygroundRequestCost: vi.fn(),
  submitPlaygroundChat: vi.fn(),
  submitPlaygroundImage: vi.fn(),
}))
vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ user: { id: 7 } }),
}))
vi.mock('@/components/layout/AppLayout.vue', () => ({
  default: { template: '<main><slot /></main>' },
}))
vi.mock('@/components/common/Select.vue', () => ({
  default: {
    name: 'Select',
    inheritAttrs: false,
    props: ['modelValue', 'options', 'placeholder'],
    emits: ['update:modelValue'],
    methods: {
      coerce(value: string) {
        if (value === '') return null
        const numeric = Number(value)
        return Number.isNaN(numeric) ? value : numeric
      },
    },
    template: `
      <select
        v-bind="$attrs"
        :value="modelValue ?? ''"
        @change="$emit('update:modelValue', coerce($event.target.value))"
      >
        <option v-for="opt in options" :key="String(opt.value)" :value="opt.value">{{ opt.label }}</option>
      </select>
    `,
  },
}))
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => (params ? `${key}:${JSON.stringify(params)}` : key),
    }),
  }
})

function session(groups: PlaygroundSession['groups']): PlaygroundSession {
  return {
    balance: 10,
    frozen_balance: 0,
    available_balance: 10,
    simple_mode: false,
    subscriptions: [],
    groups,
    limits: { max_documents: 3, max_document_bytes: 1024, max_document_runes: 1000 },
  }
}

const mounted: ReturnType<typeof mount>[] = []
function render() {
  const wrapper = mount(UseView, {
    global: {
      stubs: {
        Icon: true,
        'i18n-t': { template: '<span><slot name="handbookLink" /></span>' },
        'router-link': { template: '<a><slot /></a>' },
      },
    },
  })
  mounted.push(wrapper)
  return wrapper
}

beforeEach(() => {
  vi.clearAllMocks()
  localStorage.clear()
})

afterEach(() => {
  mounted.splice(0).forEach((wrapper) => wrapper.unmount())
})

describe('UseView playground models', () => {
  it('shows the latest group models after a session refresh', async () => {
    getPlaygroundSession
      .mockResolvedValueOnce(session([
        {
          id: 1,
          name: 'openai',
          platform: 'openai',
          rate_multiplier: 1,
          subscription_type: false,
          allow_image_generation: true,
          supports_image: true,
          has_remaining_quota: false,
          models: ['gpt-4o'],
          model_kinds: { 'gpt-4o': ['text'] },
        },
      ]))
      .mockResolvedValueOnce(session([
        {
          id: 1,
          name: 'openai',
          platform: 'openai',
          rate_multiplier: 1,
          subscription_type: false,
          allow_image_generation: true,
          supports_image: true,
          has_remaining_quota: false,
          models: ['gpt-4o', 'gpt-4o-mini'],
          model_kinds: { 'gpt-4o': ['text'], 'gpt-4o-mini': ['text'] },
        },
      ]))

    const wrapper = render()
    await flushPromises()
    const modelSelect = wrapper.get('[data-testid="playground-chat-model-select"]')
    expect(modelSelect.text()).toContain('gpt-4o')
    expect(modelSelect.text()).not.toContain('gpt-4o-mini')
    expect(getPlaygroundSession).toHaveBeenCalledTimes(1)

    await wrapper.get('[data-testid="playground-refresh-session"]').trigger('click')
    await flushPromises()
    expect(getPlaygroundSession).toHaveBeenCalledTimes(2)
    expect(modelSelect.text()).toContain('gpt-4o-mini')
  })

  it('reloads the session when switching groups so the list uses current models', async () => {
    getPlaygroundSession
      .mockResolvedValueOnce(session([
        {
          id: 1,
          name: 'one',
          platform: 'openai',
          rate_multiplier: 1,
          subscription_type: false,
          allow_image_generation: true,
          supports_image: true,
          has_remaining_quota: false,
          models: ['gpt-4o'],
          model_kinds: { 'gpt-4o': ['text'] },
        },
        {
          id: 2,
          name: 'two',
          platform: 'openai',
          rate_multiplier: 1,
          subscription_type: false,
          allow_image_generation: true,
          supports_image: true,
          has_remaining_quota: false,
          models: ['legacy-model'],
          model_kinds: { 'legacy-model': ['text'] },
        },
      ]))
      .mockResolvedValueOnce(session([
        {
          id: 1,
          name: 'one',
          platform: 'openai',
          rate_multiplier: 1,
          subscription_type: false,
          allow_image_generation: true,
          supports_image: true,
          has_remaining_quota: false,
          models: ['gpt-4o'],
          model_kinds: { 'gpt-4o': ['text'] },
        },
        {
          id: 2,
          name: 'two',
          platform: 'openai',
          rate_multiplier: 1,
          subscription_type: false,
          allow_image_generation: true,
          supports_image: true,
          has_remaining_quota: false,
          models: ['gpt-4o-mini', 'dall-e-3'],
          model_kinds: { 'gpt-4o-mini': ['text'], 'dall-e-3': ['image'] },
        },
      ]))

    const wrapper = render()
    await flushPromises()
    expect(getPlaygroundSession).toHaveBeenCalledTimes(1)
    expect(wrapper.get('[data-testid="playground-chat-model-select"]').text()).toContain('gpt-4o')

    await wrapper.get('[data-testid="playground-group-select"]').setValue(2)
    await flushPromises()
    expect(getPlaygroundSession).toHaveBeenCalledTimes(2)
    expect(wrapper.get('[data-testid="playground-chat-model-select"]').text()).toContain('gpt-4o-mini')
    expect(wrapper.get('[data-testid="playground-chat-model-select"]').text()).not.toContain('legacy-model')
  })

  it('keeps the newest session when an older refresh settles later', async () => {
    const staleGroup = {
      id: 1,
      name: 'openai',
      platform: 'openai',
      rate_multiplier: 1,
      subscription_type: false,
      allow_image_generation: true,
      supports_image: true,
      has_remaining_quota: false,
      models: ['legacy-model'],
      model_kinds: { 'legacy-model': ['text'] },
    }
    const freshGroup = {
      ...staleGroup,
      models: ['gpt-4o-mini'],
      model_kinds: { 'gpt-4o-mini': ['text'] },
    }

    let resolveStale!: (value: PlaygroundSession) => void
    getPlaygroundSession
      .mockResolvedValueOnce(session([staleGroup]))
      .mockImplementationOnce(() => new Promise((resolve) => {
        resolveStale = resolve
      }))
      .mockResolvedValueOnce(session([freshGroup]))

    const wrapper = render()
    await flushPromises()
    expect(wrapper.get('[data-testid="playground-chat-model-select"]').text()).toContain('legacy-model')

    Object.defineProperty(document, 'visibilityState', { configurable: true, value: 'visible' })
    document.dispatchEvent(new Event('visibilitychange'))
    document.dispatchEvent(new Event('visibilitychange'))
    await flushPromises()

    expect(getPlaygroundSession).toHaveBeenCalledTimes(3)
    expect(wrapper.get('[data-testid="playground-chat-model-select"]').text()).toContain('gpt-4o-mini')
    expect(wrapper.get('[data-testid="playground-chat-model-select"]').text()).not.toContain('legacy-model')

    resolveStale(session([staleGroup]))
    await flushPromises()
    expect(wrapper.get('[data-testid="playground-chat-model-select"]').text()).toContain('gpt-4o-mini')
    expect(wrapper.get('[data-testid="playground-chat-model-select"]').text()).not.toContain('legacy-model')
  })
})

describe('UseView mode tabs', () => {
  function baseSession() {
    return session([
      {
        id: 1,
        name: 'openai',
        platform: 'openai',
        rate_multiplier: 1,
        subscription_type: false,
        allow_image_generation: true,
        supports_image: true,
        has_remaining_quota: false,
        models: ['gpt-4o', 'dall-e-3'],
        model_kinds: { 'gpt-4o': ['text'], 'dall-e-3': ['image'] },
      },
    ])
  }

  it('highlights the chat tab by default and switches the selected state to image', async () => {
    getPlaygroundSession.mockResolvedValue(baseSession())

    const wrapper = render()
    await flushPromises()

    const chatTab = wrapper.get('[data-testid="playground-tab-chat"]')
    const imageTab = wrapper.get('[data-testid="playground-tab-image"]')

    expect(chatTab.text()).toBe('playground.tabChat')
    expect(imageTab.text()).toBe('playground.tabImage')
    expect(chatTab.attributes('role')).toBe('tab')
    expect(imageTab.attributes('role')).toBe('tab')
    expect(chatTab.attributes('aria-selected')).toBe('true')
    expect(imageTab.attributes('aria-selected')).toBe('false')
    expect(chatTab.classes()).toContain('playground-mode-tab')
    expect(chatTab.classes()).toContain('playground-mode-tab-active')
    expect(imageTab.classes()).toContain('playground-mode-tab')
    expect(imageTab.classes()).not.toContain('playground-mode-tab-active')
    expect(wrapper.get('[role="tablist"]').classes()).toContain('playground-mode-tabs')
    expect(wrapper.find('[data-testid="playground-chat-model-select"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="playground-image-model-select"]').exists()).toBe(false)

    await imageTab.trigger('click')

    expect(imageTab.attributes('aria-selected')).toBe('true')
    expect(chatTab.attributes('aria-selected')).toBe('false')
    expect(imageTab.classes()).toContain('playground-mode-tab-active')
    expect(chatTab.classes()).not.toContain('playground-mode-tab-active')
    expect(wrapper.find('[data-testid="playground-image-model-select"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="playground-chat-model-select"]').exists()).toBe(false)

    await chatTab.trigger('click')

    expect(chatTab.attributes('aria-selected')).toBe('true')
    expect(imageTab.attributes('aria-selected')).toBe('false')
    expect(chatTab.classes()).toContain('playground-mode-tab-active')
    expect(imageTab.classes()).not.toContain('playground-mode-tab-active')
  })
})
