import { Link, useNavigate, useOutletContext } from 'react-router'
import { applications } from '../../catalog/applications'
import type { AuthOutletContext } from '../../components/layout/AppLayout'
import { setViewMode } from '../../lib/viewMode'

export function LighthousePage() {
  const { logout } = useOutletContext<AuthOutletContext>()
  const navigate = useNavigate()
  function useClassicView() {
    setViewMode('classic')
    navigate('/overview')
  }

  return (
    <main className="lighthouse-page">
      <span className="eyebrow">ATALAYA / FARO</span>
      <div className="lighthouse-mark" aria-hidden="true">
        <span />
      </div>
      <h1>Producción, a la vista.</h1>
      <p>Elegí un horizonte para entrar al puesto de observación.</p>
      <nav aria-label="Aplicaciones">
        {applications.map((app) => (
          <Link to={`/apps/${app.slug}`} key={app.slug}>
            {app.displayName}
          </Link>
        ))}
      </nav>
      <div className="lighthouse-actions">
        <button onClick={useClassicView}>Vista clásica</button>
        <button onClick={() => void logout()}>Cerrar sesión</button>
      </div>
    </main>
  )
}
