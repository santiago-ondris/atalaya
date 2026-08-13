import { describe, expect, it } from 'vitest'
import { applications } from '../../catalog/applications'
import { findDestination, lighthouseDestinations } from './destinations'

describe('lighthouse destinations', () => {
  it('maps exactly five applications and four operational pages', () => {
    expect(lighthouseDestinations).toHaveLength(9)
    expect(lighthouseDestinations.map(({ route }) => route)).toEqual([
      ...applications.map((application) => `/apps/${application.slug}`),
      '/events',
      '/operations',
      '/reports',
      '/system',
    ])
    expect(new Set(lighthouseDestinations.map(({ id }) => id)).size).toBe(9)
  })

  it('resolves every hotspot from its stable identifier', () => {
    for (const destination of lighthouseDestinations) {
      expect(findDestination(destination.id)).toEqual(destination)
    }
  })
})
