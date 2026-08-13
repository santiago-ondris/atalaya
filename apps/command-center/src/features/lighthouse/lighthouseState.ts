import type { PublicStatus, SystemHealth } from '../../api'
import { applications, type ApplicationSlug } from '../../catalog/applications'

export type VisualSeverity = 'green' | 'yellow' | 'red'
export type Freshness = 'loading' | 'fresh' | 'stale'
export type WeatherLevel = 'clear' | 'mist' | 'storm'

export interface ApplicationVisualState {
  slug: ApplicationSlug
  severity: VisualSeverity
  freshness: Freshness
}

export interface LighthouseVisualState {
  applications: Record<ApplicationSlug, ApplicationVisualState>
  system: { severity: VisualSeverity; freshness: Freshness }
  weather: WeatherLevel
}

export interface RetainedSource<T> {
  data?: T
  freshness: Freshness
}

export interface LighthouseSources {
  publicStatus: RetainedSource<PublicStatus>
  systemHealth: RetainedSource<SystemHealth>
}

export const STALE_AFTER_MS = 2 * 60 * 1000

export function publicSeverity(status: PublicStatus['status']): VisualSeverity {
  if (status === 'operational') return 'green'
  if (status === 'major_outage') return 'red'
  return 'yellow'
}

export function systemSeverity(health: SystemHealth): VisualSeverity {
  const signalsHealthy = health.signals.every((signal) => signal.status === 'healthy')
  return health.status === 'healthy' && signalsHealthy ? 'green' : 'yellow'
}

export function worstSeverity(...values: VisualSeverity[]): VisualSeverity {
  if (values.includes('red')) return 'red'
  if (values.includes('yellow')) return 'yellow'
  return 'green'
}

export function weatherForSeverity(severity: VisualSeverity): WeatherLevel {
  return severity === 'red' ? 'storm' : severity === 'yellow' ? 'mist' : 'clear'
}

export function freshnessForGeneratedAt(
  generatedAt: string,
  now = Date.now(),
): Freshness {
  const timestamp = Date.parse(generatedAt)
  return Number.isFinite(timestamp) && now - timestamp <= STALE_AFTER_MS
    ? 'fresh'
    : 'stale'
}

export function initialSources(): LighthouseSources {
  return {
    publicStatus: { freshness: 'loading' },
    systemHealth: { freshness: 'loading' },
  }
}

export function retainSuccess<T extends { generated_at: string }>(
  data: T,
  now = Date.now(),
): RetainedSource<T> {
  return { data, freshness: freshnessForGeneratedAt(data.generated_at, now) }
}

export function retainFailure<T>(source: RetainedSource<T>): RetainedSource<T> {
  return { ...source, freshness: 'stale' }
}

export function deriveVisualState(sources: LighthouseSources): LighthouseVisualState {
  const publicBySlug = new Map(
    sources.publicStatus.data?.applications.map((application) => [
      application.slug,
      publicSeverity(application.status),
    ]),
  )
  const internalSeverity = sources.systemHealth.data
    ? systemSeverity(sources.systemHealth.data)
    : 'yellow'
  const normalized = Object.fromEntries(
    applications.map((application) => {
      const publicValue = publicBySlug.get(application.slug) ?? 'yellow'
      const severity =
        application.slug === 'atalaya'
          ? worstSeverity(publicValue, internalSeverity)
          : publicValue
      const freshness =
        application.slug === 'atalaya'
          ? combineFreshness(
              sources.publicStatus.freshness,
              sources.systemHealth.freshness,
            )
          : sources.publicStatus.freshness
      return [application.slug, { slug: application.slug, severity, freshness }]
    }),
  ) as Record<ApplicationSlug, ApplicationVisualState>
  const worst = worstSeverity(
    ...Object.values(normalized).map((application) => application.severity),
  )
  return {
    applications: normalized,
    system: {
      severity: internalSeverity,
      freshness: sources.systemHealth.freshness,
    },
    weather: weatherForSeverity(worst),
  }
}

function combineFreshness(...values: Freshness[]): Freshness {
  if (values.includes('stale')) return 'stale'
  if (values.includes('loading')) return 'loading'
  return 'fresh'
}
