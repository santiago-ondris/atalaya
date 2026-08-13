import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter, Outlet, Route, Routes } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'

vi.mock('./scene/LighthouseScene', () => ({
  LighthouseScene: () => <div data-testid="scene" />,
}))
vi.mock('./useLighthouseHealth', () => ({
  demoWeather: () => null,
  useLighthouseHealth: () => ({
    weather: 'mist',
    system: { severity: 'yellow', freshness: 'stale' },
    applications: {
      farmami: { slug: 'farmami', severity: 'green', freshness: 'fresh' },
      wheels_house: { slug: 'wheels_house', severity: 'yellow', freshness: 'stale' },
      prensap: { slug: 'prensap', severity: 'red', freshness: 'fresh' },
      notizap: { slug: 'notizap', severity: 'green', freshness: 'fresh' },
      atalaya: { slug: 'atalaya', severity: 'yellow', freshness: 'stale' },
    },
  }),
}))

import { LighthouseCover, LighthousePage, SceneBoundary } from './LighthousePage'

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

  it('provides nine equivalent DOM links in narrative order', async () => {
    render(
      <MemoryRouter>
        <Routes>
          <Route element={<Outlet context={{ logout: vi.fn() }} />}>
            <Route path="*" element={<LighthousePage />} />
          </Route>
        </Routes>
      </MemoryRouter>,
    )

    const navigation = await screen.findByRole('navigation', {
      name: 'Destinos del faro',
    })
    const links = screen.getAllByRole('link')
    expect(navigation).toContainElement(links[0])
    expect(links).toHaveLength(9)
    expect(links.map((link) => link.getAttribute('href'))).toEqual([
      '/apps/farmami',
      '/apps/wheels_house',
      '/apps/prensap',
      '/apps/notizap',
      '/apps/atalaya',
      '/events',
      '/operations',
      '/reports',
      '/system',
    ])
    expect(screen.getByRole('link', { name: 'Farmami: operativo' })).toBeVisible()
    expect(
      screen.getByRole('link', {
        name: 'Wheels House: estado degradado o desconocido, datos desactualizados',
      }),
    ).toBeVisible()
    expect(screen.getByRole('link', { name: 'Eventos' })).toHaveAccessibleName('Eventos')

    fireEvent.focus(screen.getByRole('link', { name: 'Eventos' }))
    expect(screen.getByText('Eventos', { selector: '.destination-label' })).toBeVisible()
    fireEvent.blur(screen.getByRole('link', { name: 'Eventos' }))
    expect(screen.queryByText('Eventos', { selector: '.destination-label' })).toBeNull()
  })
})
