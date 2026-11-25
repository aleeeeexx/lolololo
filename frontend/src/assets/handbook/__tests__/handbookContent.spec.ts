import { readFileSync, readdirSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

import handbookZh from '../handbook.zh.html?raw'

const handbookDir = join(process.cwd(), 'src/assets/handbook')
const publicHandbookDir = join(process.cwd(), 'public/handbook')

const banned = [
  'codecc',
  'codecc.cc',
  'www.codecc',
  'junfeng530',
  'j530274949',
  'codecc-imagegen',
]

describe('handbook brand migration', () => {
  it('keeps the Chinese handbook free of old-site brand, domains, and contacts', () => {
    const text = handbookZh.toLowerCase()
    for (const needle of banned) {
      expect(text, `found leftover "${needle}"`).not.toContain(needle)
    }
    expect(handbookZh).toContain('{{API_BASE}}')
    expect(handbookZh).toContain('DubDub')
    expect(handbookZh).toContain('/keys')
    expect(handbookZh).toContain('/register')
    expect(handbookZh).toContain('/use')
    expect(handbookZh).toContain('/handbook/images/')
    expect(handbookZh).not.toContain('https://codecc')
  })

  it('hosts screenshots and the image skill locally instead of hotlinking the old site', () => {
    const html = readFileSync(join(handbookDir, 'handbook.zh.html'), 'utf8')
    expect(html).not.toMatch(/https?:\/\/[^"' ]*codecc/i)
    const images = readdirSync(join(publicHandbookDir, 'images'))
    expect(images).toEqual(expect.arrayContaining([
      'img-01.png',
      'img-16-ccswitch-codex-v4.png',
      'img-17-ccswitch-claude-v4.png',
    ]))
    const zip = readFileSync(join(publicHandbookDir, 'dubdub-imagegen.zip'))
    expect(zip.subarray(0, 2).toString()).toBe('PK')
  })

  it('puts short quick-start examples before account registration', () => {
    const quickAt = handbookZh.indexOf('id="quick-examples"')
    const registerAt = handbookZh.indexOf('id="register"')
    expect(quickAt).toBeGreaterThan(0)
    expect(quickAt).toBeLessThan(registerAt)
    expect(handbookZh).toContain('id="quick-ex-gpt"')
    expect(handbookZh).toContain('不要登录 Codex')
    expect(handbookZh).toContain('重启 Codex')
    expect(handbookZh).toContain('id="quick-ex-claude"')
    expect(handbookZh).toContain('id="quick-ex-use"')
    expect(handbookZh).toContain('href="/use"')
    expect(handbookZh).toContain('href="#keys"')
    expect(handbookZh).toContain('href="#download"')
    expect(handbookZh).toContain('href="#config-gpt"')
    expect(handbookZh).toContain('href="#claude-ccswitch"')
    expect(handbookZh).toContain('href="#use-client"')
    expect(handbookZh).toContain('带有<code>gpt-</code>的')
    expect(handbookZh).toContain('带有<code>claude-</code>的')
    expect(handbookZh).toContain('再运行<code>claude</code>。')
    expect(handbookZh).toContain('Codex 官网')
    expect(handbookZh).toContain('后文「Codex 客户端」')
    expect(handbookZh).not.toContain('详见Codex')
    expect(handbookZh).not.toContain('详见<a href="#use-client">Codex 客户端</a>')
  })
})
