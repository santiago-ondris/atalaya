import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'
import { CommandPaletteProvider, CommandPaletteTrigger } from './CommandPalette'

afterEach(cleanup)

function Harness() {
  const location = useLocation()
  return (
    <CommandPaletteProvider>
      <button type="button">Punto de retorno</button>
      <CommandPaletteTrigger />
      <output aria-label="Ruta actual">{location.pathname}</output>
    </CommandPaletteProvider>
  )
}

function renderPalette() {
  return render(
    <MemoryRouter initialEntries={['/events']}>
      <Routes>
        <Route path="*" element={<Harness />} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('command palette', () => {
  it('opens from its trigger, focuses search, and restores trigger focus on Escape', async () => {
    renderPalette()
    const trigger = screen.getByRole('button', { name: /Ir a/ })
    fireEvent.click(trigger)

    const input = await screen.findByRole('combobox', { name: 'Buscar destino' })
    await waitFor(() => expect(input).toHaveFocus())
    expect(screen.getAllByRole('option')).toHaveLength(9)

    fireEvent.keyDown(input, { key: 'Escape' })
    await waitFor(() => expect(trigger).toHaveFocus())
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it.each([
    { modifier: { metaKey: true }, label: 'Meta+K' },
    { modifier: { ctrlKey: true }, label: 'Ctrl+K' },
  ])(
    'opens with $label and returns focus to the keyboard origin',
    async ({ modifier }) => {
      renderPalette()
      const origin = screen.getByRole('button', { name: 'Punto de retorno' })
      origin.focus()

      fireEvent.keyDown(document, { key: 'k', ...modifier })
      const input = await screen.findByRole('combobox', { name: 'Buscar destino' })
      await waitFor(() => expect(input).toHaveFocus())
      fireEvent.keyDown(input, { key: 'Escape' })
      await waitFor(() => expect(origin).toHaveFocus())
    },
  )

  it('does not open for K without a modifier', () => {
    renderPalette()
    fireEvent.keyDown(document, { key: 'k' })
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('closes on an outside interaction and restores the opener', async () => {
    renderPalette()
    const trigger = screen.getByRole('button', { name: /Ir a/ })
    fireEvent.click(trigger)
    await screen.findByRole('dialog', { name: 'Buscar destino' })
    const overlay = document.querySelector('[cmdk-overlay]')
    expect(overlay).not.toBeNull()
    fireEvent.pointerDown(overlay!)
    fireEvent.click(overlay!)
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    await waitFor(() => expect(trigger).toHaveFocus())
  })

  it.each([
    ['prensapp', 'Prensap'],
    ['wheelshouse', 'Wheels House'],
    ['watchman', 'Atalaya'],
    ['errores', 'Eventos'],
    ['despliegues', 'Bitácora'],
    ['informes', 'Reportes'],
    ['salud', 'Estado del sistema'],
    ['colas', 'Bitácora'],
  ])('filters %s to %s', async (query, destination) => {
    renderPalette()
    fireEvent.click(screen.getByRole('button', { name: /Ir a/ }))
    const input = await screen.findByRole('combobox', { name: 'Buscar destino' })
    fireEvent.change(input, { target: { value: query } })

    await waitFor(() => {
      const visibleOptions = screen
        .getAllByRole('option', { hidden: true })
        .filter((option) => !option.hasAttribute('hidden'))
      expect(visibleOptions).toHaveLength(1)
      expect(visibleOptions[0]).toHaveTextContent(destination)
    })
  })

  it('announces an empty state and navigates with arrows and Enter', async () => {
    renderPalette()
    fireEvent.click(screen.getByRole('button', { name: /Ir a/ }))
    const input = await screen.findByRole('combobox', { name: 'Buscar destino' })
    fireEvent.change(input, { target: { value: 'sin-coincidencias' } })
    expect(await screen.findByText('Sin resultados')).toBeVisible()

    fireEvent.change(input, { target: { value: 'reportes' } })
    fireEvent.keyDown(input, { key: 'ArrowDown' })
    fireEvent.keyDown(input, { key: 'Enter' })
    await waitFor(() =>
      expect(screen.getByLabelText('Ruta actual')).toHaveTextContent('/reports'),
    )
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })
})
