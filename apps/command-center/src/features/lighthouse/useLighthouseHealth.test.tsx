import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { PublicStatus } from '../../api'
import { LIGHTHOUSE_POLL_MS, useLighthouseHealth } from './useLighthouseHealth'

const generated_at = new Date().toISOString()
const publicStatus: PublicStatus = {
  status: 'operational',
  generated_at,
  incidents: [],
  applications: [
    {
      slug: 'farmami',
      display_name: 'Farmami',
      status: 'operational',
      components: [],
      uptime_30_days: null,
    },
  ],
}
afterEach(() => {
  vi.useRealTimers()
  vi.restoreAllMocks()
})

describe('useLighthouseHealth', () => {
  it('fetches both sources in parallel and retains independent results', async () => {
    const client = {
      publicStatus: vi.fn().mockResolvedValue(publicStatus),
      systemHealth: vi.fn().mockRejectedValue(new Error('offline')),
    }
    const { result } = renderHook(() => useLighthouseHealth(client))
    expect(client.publicStatus).toHaveBeenCalledOnce()
    expect(client.systemHealth).toHaveBeenCalledOnce()
    await waitFor(() =>
      expect(result.current.applications.farmami.severity).toBe('green'),
    )
    expect(result.current.system.freshness).toBe('stale')
  })

  it('polls at 30 seconds and aborts active requests on unmount', () => {
    vi.useFakeTimers()
    const signals: AbortSignal[] = []
    const pending = (signal?: AbortSignal) => {
      if (signal) signals.push(signal)
      return new Promise<never>(() => undefined)
    }
    const client = { publicStatus: vi.fn(pending), systemHealth: vi.fn(pending) }
    const { unmount } = renderHook(() => useLighthouseHealth(client))
    expect(client.publicStatus).toHaveBeenCalledTimes(1)
    act(() => vi.advanceTimersByTime(LIGHTHOUSE_POLL_MS))
    expect(client.publicStatus).toHaveBeenCalledTimes(2)
    expect(signals[0].aborted).toBe(true)
    unmount()
    expect(signals.every((signal) => signal.aborted)).toBe(true)
    act(() => vi.advanceTimersByTime(LIGHTHOUSE_POLL_MS))
    expect(client.publicStatus).toHaveBeenCalledTimes(2)
  })
})
