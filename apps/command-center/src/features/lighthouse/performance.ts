export type SceneQuality = 'normal' | 'reduced'
export type PerformanceSimulation = 'degraded' | 'critical' | null

export interface PerformanceState {
  quality: SceneQuality
  fps: number
  lowWindows: number
  criticalWindows: number
  action: 'none' | 'classic'
}

export const PERFORMANCE_LIMITS = {
  warmupMs: 3_000,
  windowMs: 1_000,
  normalFps: 50,
  criticalFps: 30,
  lowWindows: 3,
  criticalWindows: 10,
} as const

export const QUALITY_PROFILES = {
  normal: {
    dpr: 1.5,
    shadows: true,
    shadowMap: 1024,
    seaSegments: 48,
    clouds: 5,
    stars: 90,
    boats: 6,
  },
  reduced: {
    dpr: 1,
    shadows: false,
    shadowMap: 0,
    seaSegments: 24,
    clouds: 3,
    stars: 45,
    boats: 2,
  },
} as const satisfies Record<SceneQuality, object>

export function initialPerformanceState(): PerformanceState {
  return { quality: 'normal', fps: 0, lowWindows: 0, criticalWindows: 0, action: 'none' }
}

export function evaluatePerformanceWindow(
  state: PerformanceState,
  measuredFps: number,
): PerformanceState {
  if (state.action === 'classic') return state
  if (state.quality === 'normal') {
    const lowWindows =
      measuredFps < PERFORMANCE_LIMITS.normalFps ? state.lowWindows + 1 : 0
    return {
      ...state,
      fps: measuredFps,
      lowWindows,
      quality: lowWindows >= PERFORMANCE_LIMITS.lowWindows ? 'reduced' : 'normal',
      criticalWindows: 0,
    }
  }
  const criticalWindows =
    measuredFps < PERFORMANCE_LIMITS.criticalFps ? state.criticalWindows + 1 : 0
  return {
    ...state,
    fps: measuredFps,
    criticalWindows,
    action: criticalWindows >= PERFORMANCE_LIMITS.criticalWindows ? 'classic' : 'none',
  }
}

export function simulatedFps(simulation: PerformanceSimulation, actual: number) {
  return simulation === 'degraded' ? 45 : simulation === 'critical' ? 20 : actual
}

export class FrameWindowMonitor {
  private startedAt: number | null = null
  private windowAt: number | null = null
  private frames = 0
  private hiddenAt: number | null = null

  frame(now: number, visible: boolean): number | null {
    if (!visible) {
      this.hiddenAt ??= now
      this.windowAt = null
      this.frames = 0
      return null
    }
    if (this.hiddenAt !== null) {
      if (this.startedAt !== null) this.startedAt += now - this.hiddenAt
      this.hiddenAt = null
    }
    this.startedAt ??= now
    if (now - this.startedAt < PERFORMANCE_LIMITS.warmupMs) return null
    this.windowAt ??= now
    this.frames += 1
    const elapsed = now - this.windowAt
    if (elapsed < PERFORMANCE_LIMITS.windowMs) return null
    const fps = (this.frames * 1_000) / elapsed
    this.windowAt = now
    this.frames = 0
    return fps
  }
}
