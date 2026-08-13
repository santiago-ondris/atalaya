import { useEffect, useRef, useState } from 'react'
import { NavLink } from 'react-router'
import { applications, type Application } from '../../catalog/applications'

export function ArchitecturePage({ app: activeApp }: { app: Application }) {
  const [isLoading, setIsLoading] = useState<boolean>(true)
  const [reloadKey, setReloadKey] = useState<number>(0)
  const [isFullscreen, setIsFullscreen] = useState<boolean>(false)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => setIsLoading(true), [activeApp.slug])

  function handleReload() {
    setIsLoading(true)
    setReloadKey((prev) => prev + 1)
  }

  function toggleFullscreen() {
    setIsFullscreen((prev) => !prev)
  }

  return (
    <main className={`page architecture-page ${isFullscreen ? 'is-fullscreen' : ''}`}>
      <header className="page-title">
        <div>
          <span className="eyebrow">05 / ARQUITECTURA ESTRUCTURAL</span>
          <h1>Diagramas de sistema</h1>
          <p>Visualización interactiva de flujos, componentes y observabilidad.</p>
        </div>

        <div className="architecture-header-actions">
          <a
            className="architecture-action-btn secondary"
            href={activeApp.htmlUrl}
            target="_blank"
            rel="noreferrer"
            title="Abrir diagrama en ventana independiente"
          >
            <i className="ti ti-external-link" />
            <span>Pestaña nueva</span>
          </a>

          <button
            className="architecture-action-btn secondary"
            onClick={handleReload}
            title="Recargar vista interactiva"
            type="button"
          >
            <i className="ti ti-refresh" />
            <span>Recargar</span>
          </button>

          <button
            className={`architecture-action-btn ${isFullscreen ? 'active' : 'primary'}`}
            onClick={toggleFullscreen}
            title={isFullscreen ? 'Salir de pantalla completa' : 'Pantalla completa'}
            type="button"
          >
            <i
              className={`ti ${isFullscreen ? 'ti-arrows-minimize' : 'ti-arrows-maximize'}`}
            />
            <span>{isFullscreen ? 'Restaurar' : 'Pantalla completa'}</span>
          </button>
        </div>
      </header>

      {/* Navigation tabs for apps */}
      <nav
        aria-label="Selección de arquitectura por aplicación"
        className="architecture-tabs"
      >
        {applications.map((app) => {
          const isSelected = app.slug === activeApp.slug
          return (
            <NavLink
              className={`arch-tab ${isSelected ? 'active' : ''}`}
              key={app.slug}
              to={`/architecture/${app.slug}`}
            >
              <span
                className="arch-tab-color-dot"
                style={{ backgroundColor: app.brandColor }}
              />
              <span className="arch-tab-name">{app.displayName}</span>
              <span className="arch-tab-badge">{app.badge}</span>
            </NavLink>
          )
        })}
      </nav>

      {/* App metadata bar */}
      <div className="architecture-meta-bar">
        <div className="arch-meta-info">
          <span className="arch-app-title">
            <span
              className="arch-dot-indicator"
              style={{ backgroundColor: activeApp.brandColor }}
            />
            {activeApp.displayName}
          </span>
          <span className="arch-meta-pill">
            <i className="ti ti-stack-2" />
            {activeApp.stack}
          </span>
          <span className="arch-meta-pill">
            <i className="ti ti-activity" />
            {activeApp.badge}
          </span>
          <span className="arch-meta-pill">
            <i className="ti ti-cloud-upload" />
            {activeApp.deploy}
          </span>
        </div>
      </div>

      {/* Diagram iframe viewport */}
      <div className="architecture-viewport-card" ref={containerRef}>
        {isLoading && (
          <div className="architecture-loading-overlay">
            <div className="arch-spinner" />
            <span>Cargando diagrama de {activeApp.displayName}...</span>
          </div>
        )}

        <iframe
          className="architecture-iframe"
          key={`${activeApp.slug}-${reloadKey}`}
          onLoad={() => setIsLoading(false)}
          src={activeApp.htmlUrl}
          title={`Diagrama interactivo de arquitectura - ${activeApp.displayName}`}
        />
      </div>
    </main>
  )
}
