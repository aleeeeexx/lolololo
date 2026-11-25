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
})
