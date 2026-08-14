import { describe, expect, it } from 'vitest'
import {
  evaluatePerformanceWindow,
  FrameWindowMonitor,
  initialPerformanceState,
  QUALITY_PROFILES,
  simulatedFps,
  type PerformanceState,
} from './performance'

describe('lighthouse performance machine', () => {
  it('warms up for three seconds and pauses hidden windows', () => {
    const monitor = new FrameWindowMonitor()
    expect(monitor.frame(0, true)).toBeNull()
    expect(monitor.frame(2_999, true)).toBeNull()
    expect(monitor.frame(3_000, true)).toBeNull()
    expect(monitor.frame(4_000, false)).toBeNull()
    expect(monitor.frame(5_000, true)).toBeNull()
    expect(monitor.frame(6_000, true)).toBe(2)
  })

  it('reduces after three low windows and never automatically recovers', () => {
    let state = initialPerformanceState()
    state = evaluatePerformanceWindow(state, 49)
    state = evaluatePerformanceWindow(state, 55)
    expect(state.lowWindows).toBe(0)
    state = evaluatePerformanceWindow(state, 49)
    state = evaluatePerformanceWindow(state, 48)
    state = evaluatePerformanceWindow(state, 47)
    expect(state.quality).toBe('reduced')
    expect(evaluatePerformanceWindow(state, 60).quality).toBe('reduced')
  })

  it('falls back after ten critical reduced windows and resets on recovery', () => {
    let state: PerformanceState = { ...initialPerformanceState(), quality: 'reduced' }
    for (let index = 0; index < 6; index += 1)
      state = evaluatePerformanceWindow(state, 20)
    state = evaluatePerformanceWindow(state, 35)
    expect(state.criticalWindows).toBe(0)
    for (let index = 0; index < 10; index += 1)
      state = evaluatePerformanceWindow(state, 20)
    expect(state.action).toBe('classic')
  })

  it('provides reproducible simulations', () => {
    expect(simulatedFps('degraded', 60)).toBe(45)
    expect(simulatedFps('critical', 60)).toBe(20)
  })

  it('defines the complete normal and reduced visual budgets', () => {
    expect(QUALITY_PROFILES.normal).toEqual({
      dpr: 1.5,
      shadows: true,
      shadowMap: 1024,
      seaSegments: 48,
      clouds: 5,
      stars: 90,
      boats: 6,
    })
    expect(QUALITY_PROFILES.reduced).toEqual({
      dpr: 1,
      shadows: false,
      shadowMap: 0,
      seaSegments: 24,
      clouds: 3,
      stars: 45,
      boats: 2,
    })
  })
})
