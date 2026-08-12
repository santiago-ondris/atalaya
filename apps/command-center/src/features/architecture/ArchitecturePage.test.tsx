import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { ArchitecturePage } from './ArchitecturePage'

describe('ArchitecturePage', () => {
  afterEach(() => {
    cleanup()
  })

  it('renders architecture page with tabs and initial app selected', () => {
    render(<ArchitecturePage initialAppSlug="farmami" />)

    expect(screen.getByText('Diagramas de sistema')).toBeInTheDocument()
    expect(screen.getAllByText('Farmami').length).toBeGreaterThan(0)
    expect(screen.getByText('Vue / Laravel / MySQL')).toBeInTheDocument()
    expect(
      screen.getByTitle('Diagrama interactivo de arquitectura - Farmami'),
    ).toHaveAttribute('src', '/diagrams/arquitectura-farmami.html')
  })

  it('switches active diagram when tab is clicked', () => {
    render(<ArchitecturePage initialAppSlug="farmami" />)

    const notizapTab = screen.getByRole('button', { name: /notizap/i })
    fireEvent.click(notizapTab)

    expect(screen.getByText('Node.js / Express / Azure')).toBeInTheDocument()
    expect(
      screen.getByTitle('Diagrama interactivo de arquitectura - Notizap'),
    ).toHaveAttribute('src', '/diagrams/arquitectura-notizap.html')
  })

  it('toggles fullscreen mode when button is clicked', () => {
    render(<ArchitecturePage initialAppSlug="prensap" />)

    const fullscreenBtn = screen.getByTitle('Pantalla completa')
    fireEvent.click(fullscreenBtn)

    expect(screen.getByTitle('Salir de pantalla completa')).toBeInTheDocument()
  })
})
