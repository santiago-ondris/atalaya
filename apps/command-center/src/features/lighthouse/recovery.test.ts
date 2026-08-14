import { beforeEach, describe, expect, it } from 'vitest'
import {
  clearFallbackReason,
  getFallbackReason,
  hasWebGL,
  setFallbackReason,
} from './recovery'

describe('lighthouse recovery', () => {
  beforeEach(() => sessionStorage.clear())

  it('persists and clears a session fallback reason', () => {
    setFallbackReason('context-lost')
    expect(getFallbackReason()).toBe('context-lost')
    clearFallbackReason()
    expect(getFallbackReason()).toBeNull()
  })

  it('rejects invalid stored reasons', () => {
    sessionStorage.setItem('atalaya:lighthouse-fallback', 'unknown')
    expect(getFallbackReason()).toBeNull()
  })

  it('detects WebGL without importing the scene', () => {
    const getContext = (name: string) => (name === 'webgl2' ? {} : null)
    const fakeDocument = { createElement: () => ({ getContext }) } as unknown as Document
    expect(hasWebGL(fakeDocument)).toBe(true)
  })
})
