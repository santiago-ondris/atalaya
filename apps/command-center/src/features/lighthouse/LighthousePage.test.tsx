import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { LighthouseCover, SceneBoundary } from './LighthousePage'

afterEach(cleanup)

function BrokenScene(): never {
  throw new Error('scene failed')
}

describe('lighthouse loading and recovery', () => {
  it('shows the two-dimensional loading cover', () => {
    render(<LighthouseCover message="Preparando el horizonte…" />)
    expect(screen.getByText('Producción, a la vista.')).toBeInTheDocument()
    expect(screen.getByText('Preparando el horizonte…')).toBeInTheDocument()
  })

  it('isolates scene errors and keeps a recovery action available', () => {
    const onClassic = vi.fn()
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => undefined)
    render(
      <SceneBoundary fallback={<button onClick={onClassic}>Abrir vista clásica</button>}>
        <BrokenScene />
      </SceneBoundary>,
    )

    fireEvent.click(screen.getByRole('button', { name: 'Abrir vista clásica' }))
    expect(onClassic).toHaveBeenCalledOnce()
    consoleError.mockRestore()
  })
})
