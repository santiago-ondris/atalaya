import type { ReactNode } from 'react'

export type View = 'overview' | 'events' | 'operations' | 'reports' | 'system'

interface AppLayoutProps {
  view: View
  children: ReactNode
  onNavigate: (view: View) => void
  onLogout: () => void
}

const navigationItems: Array<{
  view: View
  label: string
  icon: string
}> = [
  { view: 'overview', label: 'Overview', icon: 'ti-layout-dashboard' },
  { view: 'events', label: 'Eventos', icon: 'ti-alert-triangle' },
  { view: 'operations', label: 'Bitácora', icon: 'ti-timeline-event' },
  { view: 'reports', label: 'Reportes', icon: 'ti-file-report' },
  { view: 'system', label: 'Sistema', icon: 'ti-activity' },
]

export function AppLayout({ view, children, onNavigate, onLogout }: AppLayoutProps) {
  const currentTime = new Intl.DateTimeFormat('es-AR', {
    timeStyle: 'medium',
  }).format(new Date())

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="wordmark">Atalaya</div>

        <nav aria-label="Navegación principal">
          {navigationItems.map((item) => (
            <button
              className={view === item.view ? 'active' : ''}
              key={item.view}
              onClick={() => onNavigate(item.view)}
            >
              <i className={`ti ${item.icon}`} />
              {item.label}
            </button>
          ))}
        </nav>

        <div className="guard">
          <span>Guardia activa</span>
          <button onClick={onLogout}>Cerrar sesión</button>
        </div>
      </aside>

      <div className="workspace">
        <header className="command-strip">
          <span>Producción / Argentina</span>
          <span>{currentTime} ART</span>
        </header>
        {children}
      </div>
    </div>
  )
}
