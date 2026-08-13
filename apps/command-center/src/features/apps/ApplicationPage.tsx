import { Link } from 'react-router'
import type { Application } from '../../catalog/applications'

export function ApplicationPage({ app }: { app: Application }) {
  return (
    <main className="page application-page">
      <span className="eyebrow">APLICACIÓN / {app.slug.toUpperCase()}</span>
      <h1>{app.displayName}</h1>
      <p>{app.stack}</p>
      <dl>
        <dt>Observabilidad</dt>
        <dd>{app.badge}</dd>
        <dt>Deploy</dt>
        <dd>{app.deploy}</dd>
      </dl>
      <Link className="primary-action" to={`/architecture/${app.slug}`}>
        Ver Arquitectura
      </Link>
    </main>
  )
}
