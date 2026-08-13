import { NavLink, Outlet, useLocation, useNavigate, useOutletContext } from 'react-router'
import { setViewMode } from '../../lib/viewMode'
import type { AuthOutletContext } from './AppLayout'

const sectionNames: Record<string, string> = {
  apps: 'Aplicaciones',
  events: 'Eventos',
  operations: 'Bitácora',
  reports: 'Reportes',
  architecture: 'Arquitectura',
  system: 'Sistema',
}

export function FineLayout() {
  const { logout } = useOutletContext<AuthOutletContext>()
  const navigate = useNavigate()
  const section = sectionNames[useLocation().pathname.split('/')[1]] ?? 'Atalaya'

  function useClassicView() {
    setViewMode('classic')
    navigate('/overview')
  }

  return (
    <div className="fine-shell">
      <header className="fine-header">
        <NavLink to="/" className="fine-home">
          ← Faro
        </NavLink>
        <span>{section}</span>
        <div>
          <button onClick={useClassicView}>Vista clásica</button>
          <button onClick={() => void logout()}>Cerrar sesión</button>
        </div>
      </header>
      <Outlet />
    </div>
  )
}
