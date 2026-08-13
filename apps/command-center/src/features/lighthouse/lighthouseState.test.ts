import { describe, expect, it } from 'vitest'
import type { PublicStatus, SystemHealth } from '../../api'
import {
  deriveVisualState,
  freshnessForGeneratedAt,
  initialSources,
  publicSeverity,
  retainFailure,
  retainSuccess,
  systemSeverity,
  weatherForSeverity,
  worstSeverity,
} from './lighthouseState'

const now = Date.parse('2026-08-13T12:00:00Z')

function publicData(statuses: Record<string, PublicStatus['status']>): PublicStatus {
  return {
    status: 'operational',
    generated_at: new Date(now).toISOString(),
    applications: Object.entries(statuses).map(([slug, status]) => ({
      slug,
      display_name: slug,
      status,
      components: [],
      uptime_30_days: null,
    })),
    incidents: [],
  }
}

function health(status = 'healthy', signal = 'healthy'): SystemHealth {
  return {
    status,
    generated_at: new Date(now).toISOString(),
    signals: [{ name: 'database', status: signal }],
    queues: {},
  }
}

describe('lighthouse visual state', () => {
  it.each([
    ['operational', 'green'],
    ['degraded', 'yellow'],
    ['unknown', 'yellow'],
    ['major_outage', 'red'],
  ] as const)('maps public %s to %s', (input, expected) => {
    expect(publicSeverity(input)).toBe(expected)
  })

  it('only considers fully healthy internal health green', () => {
    expect(systemSeverity(health())).toBe('green')
    expect(systemSeverity(health('degraded'))).toBe('yellow')
    expect(systemSeverity(health('healthy', 'unknown'))).toBe('yellow')
  })

  it('derives weather from the worst severity', () => {
    expect(worstSeverity('green', 'yellow', 'red')).toBe('red')
    expect(weatherForSeverity('green')).toBe('clear')
    expect(weatherForSeverity('yellow')).toBe('mist')
    expect(weatherForSeverity('red')).toBe('storm')
  })

  it('starts unknown yellow, retains stale data, and recovers independently', () => {
    let sources = initialSources()
    expect(deriveVisualState(sources).applications.farmami.severity).toBe('yellow')
    sources = {
      publicStatus: retainSuccess(
        publicData({
          farmami: 'operational',
          wheels_house: 'operational',
          prensap: 'operational',
          notizap: 'operational',
          atalaya: 'operational',
        }),
        now,
      ),
      systemHealth: retainSuccess(health(), now),
    }
    expect(deriveVisualState(sources).weather).toBe('clear')
    sources = { ...sources, publicStatus: retainFailure(sources.publicStatus) }
    expect(deriveVisualState(sources).applications.farmami).toMatchObject({
      severity: 'green',
      freshness: 'stale',
    })
    sources = {
      ...sources,
      publicStatus: retainSuccess(publicData({ farmami: 'major_outage' }), now),
    }
    expect(deriveVisualState(sources).applications.farmami).toMatchObject({
      severity: 'red',
      freshness: 'fresh',
    })
    expect(deriveVisualState(sources).system.freshness).toBe('fresh')
  })

  it('marks generated data older than two minutes stale', () => {
    expect(freshnessForGeneratedAt(new Date(now - 120_001).toISOString(), now)).toBe(
      'stale',
    )
    expect(freshnessForGeneratedAt(new Date(now - 120_000).toISOString(), now)).toBe(
      'fresh',
    )
  })
})
