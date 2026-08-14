import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter, Outlet, Route, Routes } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { CommandPaletteProvider } from '../command/CommandPalette'
import { AppLayout } from './AppLayout'
import { FineLayout } from './FineLayout'

afterEach(cleanup)

function renderLayout(layout: 'fine' | 'classic') {
  return render(
    <MemoryRouter initialEntries={['/events']}>
      <Routes>
        <Route element={<Outlet context={{ logout: vi.fn() }} />}>
          <Route
            element={
              <CommandPaletteProvider>
                {layout === 'fine' ? <FineLayout /> : <AppLayout />}
              </CommandPaletteProvider>
            }
          >
            <Route path="events" element={<main>Contenido</main>} />
          </Route>
        </Route>
      </Routes>
    </MemoryRouter>,
  )
}

describe('command palette shell integration', () => {
  it.each(['fine', 'classic'] as const)(
    'renders one trigger in the %s shell',
    (layout) => {
      renderLayout(layout)
      expect(screen.getByRole('button', { name: /Ir a/ })).toBeVisible()
      expect(screen.getAllByText('Contenido')).toHaveLength(1)
    },
  )
})
