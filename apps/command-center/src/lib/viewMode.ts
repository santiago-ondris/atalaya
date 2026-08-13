export type ViewMode = 'immersive' | 'classic'

const STORAGE_KEY = 'atalaya:view-mode'

export function getStoredViewMode(storage: Storage = sessionStorage): ViewMode | null {
  const value = storage.getItem(STORAGE_KEY)
  return value === 'immersive' || value === 'classic' ? value : null
}

export function setViewMode(mode: ViewMode, storage: Storage = sessionStorage) {
  storage.setItem(STORAGE_KEY, mode)
}

export function clearViewMode(storage: Storage = sessionStorage) {
  storage.removeItem(STORAGE_KEY)
}

export function getInitialViewMode(
  viewportWidth = window.innerWidth,
  hasFinePointer = window.matchMedia?.('(pointer: fine)').matches ?? false,
  storage: Storage = sessionStorage,
): ViewMode {
  return (
    getStoredViewMode(storage) ??
    (viewportWidth >= 901 && hasFinePointer ? 'immersive' : 'classic')
  )
}
