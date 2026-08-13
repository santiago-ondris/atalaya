import { describe, expect, it } from 'vitest'
import { applications, resolveApplication } from './applications'

describe('application catalog', () => {
  it('keeps canonical order and names', () => {
    expect(applications.map(({ slug, displayName }) => [slug, displayName])).toEqual([
      ['farmami', 'Farmami'],
      ['wheels_house', 'Wheels House'],
      ['prensap', 'Prensap'],
      ['notizap', 'Notizap'],
      ['atalaya', 'Atalaya'],
    ])
  })

  it.each([
    ['wheelshouse', 'wheels_house'],
    ['prensapp', 'prensap'],
    ['watchman', 'atalaya'],
    ['interpreter', 'atalaya'],
    ['command_center', 'atalaya'],
  ])('resolves %s to %s', (alias, canonical) => {
    expect(resolveApplication(alias)?.slug).toBe(canonical)
  })

  it('rejects unknown slugs', () => expect(resolveApplication('unknown')).toBeUndefined())
})
