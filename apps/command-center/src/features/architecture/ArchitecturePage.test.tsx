import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { ArchitecturePage } from './ArchitecturePage'
import { MemoryRouter } from 'react-router'
import { resolveApplication } from '../../catalog/applications'

function renderArchitecture(slug: string) {
  return render(
    <MemoryRouter>
      <ArchitecturePage app={resolveApplication(slug)!} />
    </MemoryRouter>,
  )
}

describe('ArchitecturePage', () => {
  afterEach(() => {
    cleanup()
  })

  it('renders architecture page with tabs and initial app selected', () => {
    renderArchitecture('farmami')

    expect(screen.getByText('Diagramas de sistema')).toBeInTheDocument()
    expect(screen.getAllByText('Farmami').length).toBeGreaterThan(0)
    expect(
      screen.getByText('Node / Prisma / Express / React / PostgreSQL'),
    ).toBeInTheDocument()
    expect(
      screen.getByTitle('Diagrama interactivo de arquitectura - Farmami'),
    ).toHaveAttribute('src', '/diagrams/arquitectura-farmami.html')
  })

  it('links every diagram tab to its canonical route', () => {
    renderArchitecture('farmami')

    expect(screen.getByRole('link', { name: /notizap/i })).toHaveAttribute(
      'href',
      '/architecture/notizap',
    )
  })

  it('toggles fullscreen mode when button is clicked', () => {
    renderArchitecture('prensap')

    const fullscreenBtn = screen.getByTitle('Pantalla completa')
    fireEvent.click(fullscreenBtn)

    expect(screen.getByTitle('Salir de pantalla completa')).toBeInTheDocument()
  })
})
