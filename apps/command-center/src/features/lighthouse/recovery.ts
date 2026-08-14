export const FALLBACK_REASON_KEY = 'atalaya:lighthouse-fallback'

export type LighthouseFailureReason =
  | 'webgl-unavailable'
  | 'scene-error'
  | 'context-lost'
  | 'performance'

export const failureMessages: Record<LighthouseFailureReason, string> = {
  'webgl-unavailable': 'WebGL no está disponible en este navegador o equipo.',
  'scene-error': 'La escena 3D no pudo cargarse o renderizarse.',
  'context-lost': 'Se perdió el contexto gráfico y no pudo recuperarse.',
  performance: 'El faro cambió a la vista clásica para mantener un rendimiento estable.',
}

export function hasWebGL(documentValue: Document = document) {
  try {
    const canvas = documentValue.createElement('canvas')
    return Boolean(canvas.getContext('webgl2') || canvas.getContext('webgl'))
  } catch {
    return false
  }
}

export function getFallbackReason(
  storage: Storage = sessionStorage,
): LighthouseFailureReason | null {
  const value = storage.getItem(FALLBACK_REASON_KEY)
  return value && value in failureMessages ? (value as LighthouseFailureReason) : null
}

export function setFallbackReason(
  reason: LighthouseFailureReason,
  storage: Storage = sessionStorage,
) {
  storage.setItem(FALLBACK_REASON_KEY, reason)
}

export function clearFallbackReason(storage: Storage = sessionStorage) {
  storage.removeItem(FALLBACK_REASON_KEY)
}
