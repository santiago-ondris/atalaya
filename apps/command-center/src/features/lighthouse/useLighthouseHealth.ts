import { useEffect, useMemo, useState } from 'react'
import { api, type PublicStatus, type SystemHealth } from '../../api'
import {
  deriveVisualState,
  initialSources,
  retainFailure,
  retainSuccess,
  type LighthouseSources,
  type WeatherLevel,
} from './lighthouseState'

export const LIGHTHOUSE_POLL_MS = 30_000

interface HealthApi {
  publicStatus: (signal?: AbortSignal) => Promise<PublicStatus>
  systemHealth: (signal?: AbortSignal) => Promise<SystemHealth>
}

export function useLighthouseHealth(client: HealthApi = api) {
  const [sources, setSources] = useState<LighthouseSources>(initialSources)

  useEffect(() => {
    let generation = 0
    let controllers: AbortController[] = []

    const refresh = () => {
      generation += 1
      const current = generation
      controllers.forEach((controller) => controller.abort())
      const publicController = new AbortController()
      const systemController = new AbortController()
      controllers = [publicController, systemController]

      void client.publicStatus(publicController.signal).then(
        (data) => {
          if (current === generation)
            setSources((value) => ({
              ...value,
              publicStatus: retainSuccess(data),
            }))
        },
        () => {
          if (current === generation)
            setSources((value) => ({
              ...value,
              publicStatus: retainFailure(value.publicStatus),
            }))
        },
      )
      void client.systemHealth(systemController.signal).then(
        (data) => {
          if (current === generation)
            setSources((value) => ({
              ...value,
              systemHealth: retainSuccess(data),
            }))
        },
        () => {
          if (current === generation)
            setSources((value) => ({
              ...value,
              systemHealth: retainFailure(value.systemHealth),
            }))
        },
      )
    }

    refresh()
    const timer = window.setInterval(refresh, LIGHTHOUSE_POLL_MS)
    return () => {
      generation += 1
      window.clearInterval(timer)
      controllers.forEach((controller) => controller.abort())
    }
  }, [client])

  return useMemo(() => deriveVisualState(sources), [sources])
}

export function demoWeather(value: string | null): WeatherLevel | null {
  if (!import.meta.env.DEV) return null
  return value === 'clear' || value === 'mist' || value === 'storm' ? value : null
}
