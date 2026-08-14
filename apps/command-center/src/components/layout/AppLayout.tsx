import { NavLink, Outlet, useNavigate, useOutletContext } from 'react-router'
import { setViewMode } from '../../lib/viewMode'
import { CommandPaletteTrigger } from '../command/CommandPalette'

export type AuthOutletContext = { logout: () => Promise<void> }

const navigationItems = [
  { to: '/overview', label: 'Overview', icon: 'ti-layout-dashboard' },
  { to: '/events', label: 'Eventos', icon: 'ti-alert-triangle' },
  { to: '/operations', label: 'Bitácora', icon: 'ti-timeline-event' },
  { to: '/reports', label: 'Reportes', icon: 'ti-file-report' },
  { to: '/architecture', label: 'Arquitectura', icon: 'ti-sitemap' },
  { to: '/system', label: 'Sistema', icon: 'ti-activity' },
] as const

export function AppLayout() {
  const { logout } = useOutletContext<AuthOutletContext>()
  const navigate = useNavigate()
  const currentTime = new Intl.DateTimeFormat('es-AR', { timeStyle: 'medium' }).format(
    new Date(),
  )

  function returnToLighthouse() {
    setViewMode('immersive')
    navigate('/')
  }

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="wordmark">Atalaya</div>
        <nav aria-label="Navegación principal">
          {navigationItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === '/overview'}
              className={({ isActive }) => (isActive ? 'active' : undefined)}
            >
              <i className={`ti ${item.icon}`} />
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="guard">
          <span>Guardia activa</span>
          <CommandPaletteTrigger />
          <button onClick={returnToLighthouse}>Volver al faro</button>
          <button onClick={() => void logout()}>Cerrar sesión</button>
        </div>
      </aside>
      <div className="workspace">
        <header className="command-strip">
          <span>Producción / Argentina</span>
          <span>{currentTime} ART</span>
        </header>
        <Outlet />
      </div>
    </div>
  )
}
