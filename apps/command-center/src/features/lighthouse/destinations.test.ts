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
    expect(new Set(lighthouseDestinations.map(({ route }) => route)).size).toBe(9)
    expect(
      lighthouseDestinations.filter(({ group }) => group === 'Aplicaciones'),
    ).toHaveLength(5)
    expect(
      lighthouseDestinations.filter(({ group }) => group === 'Operación'),
    ).toHaveLength(4)
  })

  it('centralizes application aliases and predictable operational search terms', () => {
    const searchable = Object.fromEntries(
      lighthouseDestinations.map(({ id, keywords }) => [id, keywords]),
    )
    expect(searchable.wheels_house).toContain('wheelshouse')
    expect(searchable.prensap).toContain('prensapp')
    expect(searchable.atalaya).toContain('watchman')
    expect(searchable.events).toContain('errores')
    expect(searchable.operations).toEqual(
      expect.arrayContaining(['operaciones', 'despliegues', 'colas']),
    )
    expect(searchable.reports).toContain('informes')
    expect(searchable.system).toContain('salud')
  })

  it('resolves every hotspot from its stable identifier', () => {
    for (const destination of lighthouseDestinations) {
      expect(findDestination(destination.id)).toEqual(destination)
    }
  })
})
