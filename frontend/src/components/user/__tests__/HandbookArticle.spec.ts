import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { defineComponent } from 'vue'
import { createI18n } from 'vue-i18n'
import { createRouter, createMemoryHistory } from 'vue-router'

import zh from '@/i18n/locales/zh'
import en from '@/i18n/locales/en'
import HandbookArticle from '../HandbookArticle.vue'

const { appStore } = vi.hoisted(() => ({
  appStore: {
    apiBaseUrl: 'https://api.dubdub-ai.cn',
    fetchPublicSettings: vi.fn().mockResolvedValue({}),
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => appStore,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStore,
}))

function mountArticle(locale: 'zh' | 'en' = 'zh') {
  const i18n = createI18n({
    legacy: false,
    locale,
    messages: { zh, en },
  })
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: defineComponent({ template: '<div />' }) },
      { path: '/handbook', component: defineComponent({ template: '<div />' }) },
      { path: '/keys', component: defineComponent({ template: '<div />' }) },
      { path: '/register', component: defineComponent({ template: '<div />' }) },
    ],
  })
  return mount(HandbookArticle, {
    global: {
      plugins: [i18n, router],
    },
  })
}

describe('HandbookArticle', () => {
  beforeEach(() => {
    appStore.apiBaseUrl = 'https://api.dubdub-ai.cn'
    Object.defineProperty(window, 'scrollTo', { value: vi.fn(), writable: true })
  })

  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('renders the DubDub handbook with the resolved API base and no old-site leftovers', async () => {
    const wrapper = mountArticle('zh')
    await wrapper.vm.$nextTick()
    const text = wrapper.text()
    const html = wrapper.html()
    expect(text).toContain('DubDub 使用手册')
    expect(text).toContain('DubDub AI中转站')
    expect(text).toContain('https://api.dubdub-ai.cn')
    expect(text).toContain('一、注册账号')
    expect(text).not.toContain('极简示例')
    expect(text).not.toContain('想用 GPT（Codex）')
    expect(text).not.toContain('不装客户端，在浏览器里直接用')
    expect(text).not.toContain('注意事项：API 请求地址不要漏掉 /v1')
    expect(wrapper.get('.handbook-toc-link').text()).toBe('一、注册账号')
    expect(wrapper.find('#quick-examples').exists()).toBe(false)
    expect(wrapper.find('#ccswitch-api-url').exists()).toBe(false)
    expect(html).not.toContain('#quick-examples')
    expect(html).not.toContain('#ccswitch-api-url')
    expect(html).toContain('/handbook/images/img-01.png')
    expect(html.toLowerCase()).not.toContain('codecc')
    expect(html).not.toContain('j530274949')
    expect(html).not.toContain('junfeng530')
    expect(html).not.toContain('{{API_BASE}}')
    wrapper.unmount()
  })

  it('shows an English notice while keeping the Chinese body', async () => {
    const wrapper = mountArticle('en')
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('This handbook is currently available in Chinese.')
    expect(wrapper.text()).toContain('DubDub 使用手册')
    expect(wrapper.get('.handbook-toc-link').text()).toBe('1. Create an account')
    expect(wrapper.find('#quick-examples').exists()).toBe(false)
    expect(wrapper.find('#ccswitch-api-url').exists()).toBe(false)
    expect(wrapper.get('#register').text()).toBe('1. Create an account')
    expect(wrapper.text()).not.toContain('Note: do not omit /v1 in the CC Switch API request address')
    expect(wrapper.html()).not.toContain('#quick-examples')
    expect(wrapper.html()).not.toContain('#ccswitch-api-url')
    wrapper.unmount()
  })

  it('places a fixed-width table of contents to the left of a flexible main column', async () => {
    const source = readFileSync(join(process.cwd(), 'src/components/user/HandbookArticle.vue'), 'utf8')
    expect(source).toMatch(/grid-template-columns:\s*minmax\(16rem,\s*18rem\)\s+minmax\(0,\s*1fr\)/)
    expect(source).not.toMatch(/grid-template-columns:\s*minmax\(0,\s*1fr\)\s+220px/)
    expect(source).not.toContain('overflow-wrap: anywhere')

    const wrapper = mountArticle('zh')
    await wrapper.vm.$nextTick()
    const root = wrapper.get('.handbook-page')
    const toc = wrapper.get('.handbook-toc')
    const main = wrapper.get('.handbook-main')
    expect(root.element.firstElementChild).toBe(toc.element)
    expect(toc.element.nextElementSibling).toBe(main.element)
    expect(wrapper.get('.handbook-api-url').text()).toBe('https://api.dubdub-ai.cn')
    expect(wrapper.get('.handbook-api-url').classes()).not.toContain('break-all')
    expect(main.text()).toContain('DubDub 使用手册')
    wrapper.unmount()
  })
})
