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

  it('does not include the removed quick-start examples section', () => {
    expect(handbookZh).not.toContain('id="quick-examples"')
    expect(handbookZh).not.toContain('id="quick-ex-gpt"')
    expect(handbookZh).not.toContain('id="quick-ex-claude"')
    expect(handbookZh).not.toContain('id="quick-ex-use"')
    expect(handbookZh).not.toContain('#quick-examples')
    expect(handbookZh).not.toContain('极简示例')
  })

  it('does not include a dedicated CC Switch /v1 address-note section', () => {
    expect(handbookZh).not.toContain('id="ccswitch-api-url"')
    expect(handbookZh).not.toContain('#ccswitch-api-url')
    expect(handbookZh).not.toContain('注意事项：API 请求地址不要漏掉 /v1')
  })

  it('leaves no #quick-examples or #ccswitch-api-url links in non-test frontend source', () => {
    const srcRoot = join(process.cwd(), 'src')
    const files: string[] = []
    const walk = (dir: string) => {
      for (const entry of readdirSync(dir, { withFileTypes: true })) {
        const full = join(dir, entry.name)
        if (entry.isDirectory()) {
          if (entry.name === '__tests__' || entry.name === 'node_modules') continue
          walk(full)
        } else if (/\.(vue|ts|tsx|js|html)$/.test(entry.name)) {
          files.push(full)
        }
      }
    }
    walk(srcRoot)
    expect(files.length).toBeGreaterThan(0)
    for (const file of files) {
      const text = readFileSync(file, 'utf8')
      expect(text, file).not.toContain('#quick-examples')
      expect(text, file).not.toContain('id="quick-examples"')
      expect(text, file).not.toContain('极简示例')
      expect(text, file).not.toContain('#ccswitch-api-url')
      expect(text, file).not.toContain('id="ccswitch-api-url"')
    }
  })
})
