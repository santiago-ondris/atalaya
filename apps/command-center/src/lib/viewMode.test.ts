import { beforeEach, describe, expect, it } from 'vitest'
import {
  clearViewMode,
  getInitialViewMode,
  getStoredViewMode,
  setViewMode,
} from './viewMode'

describe('view mode', () => {
  beforeEach(() => sessionStorage.clear())
  it('defaults by viewport and pointer', () => {
    expect(getInitialViewMode(1200, true)).toBe('immersive')
    expect(getInitialViewMode(900, true)).toBe('classic')
    expect(getInitialViewMode(1200, false)).toBe('classic')
  })
  it('persists and clears the session preference', () => {
    setViewMode('classic')
    expect(getStoredViewMode()).toBe('classic')
    expect(getInitialViewMode(1200, true)).toBe('classic')
    clearViewMode()
    expect(getStoredViewMode()).toBeNull()
  })
})
