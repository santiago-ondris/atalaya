import { useEffect, useState } from 'react'
import { api, type PublicStatus } from '../../api'
import { formatDateTime } from '../../lib/format'
import './status.css'

const labels: Record<string, string> = {
  operational: 'Todos los sistemas operativos',
  degraded: 'Servicio degradado',
  major_outage: 'Interrupción importante',
  unknown: 'Comprobando sistemas',
}

export function StatusPage() {
  const [data, setData] = useState<PublicStatus | null>(null)
  const [error, setError] = useState(false)
  useEffect(() => {
    let active = true
    const load = () =>
      api
        .publicStatus()
        .then((v) => {
          if (active) {
            setData(v)
            setError(false)
          }
        })
        .catch(() => active && setError(true))
    load()
    const timer = setInterval(load, 60_000)
    return () => {
      active = false
      clearInterval(timer)
    }
  }, [])
  return (
    <main className="public-status">
      <header className="status-mast">
        <a href="/status" className="status-brand">
          <span>ATALAYA</span>
          <small>ESTADO DE LA FLOTA</small>
        </a>
        <span className="status-live">MONITOREO EN VIVO</span>
      </header>
      <section className="status-hero">
        <p className="eyebrow">REPORTE DE NAVEGACIÓN</p>
        <h1>
          {error ? 'No pudimos obtener el estado' : labels[data?.status ?? 'unknown']}
        </h1>
        <p>
          {error
            ? 'Reintentaremos automáticamente.'
            : data
              ? `Última actualización ${formatDateTime(data.generated_at)}`
              : 'Consultando la última telemetría…'}
        </p>
      </section>
      {data && (
        <>
          <section className="status-grid" aria-label="Aplicaciones">
            {data.applications.map((app) => (
              <article className="status-card" key={app.slug}>
                <div className="status-card-head">
                  <h2>{app.display_name}</h2>
                  <span className={`status-pill ${app.status}`}>
                    {labels[app.status]}
                  </span>
                </div>
                <div className="uptime">
                  <strong>
                    {app.uptime_30_days == null
                      ? '—'
                      : `${app.uptime_30_days.toFixed(2)}%`}
                  </strong>
                  <span>uptime · 30 días</span>
                </div>
                <ul>
                  {['frontend', 'backend'].map((name) => {
                    const component = app.components.find((c) => c.name === name)
                    return (
                      <li key={name}>
                        <span>{name === 'frontend' ? 'Sitio web' : 'API'}</span>
                        <span
                          className={`component-state ${component?.status ?? 'unknown'}`}
                        >
                          {component?.status === 'up'
                            ? 'Operativo'
                            : component?.status === 'down'
                              ? 'Caído'
                              : 'Sin datos'}
                        </span>
                      </li>
                    )
                  })}
                </ul>
                <small>Chequeado {formatDateTime(app.last_checked_at)}</small>
              </article>
            ))}
          </section>
          <section className="incident-log">
            <p className="eyebrow">BITÁCORA PÚBLICA</p>
            <h2>Incidentes recientes</h2>
            {(data.incidents ?? []).length === 0 ? (
              <p className="quiet">
                No hay incidentes publicados en los últimos 30 días.
              </p>
            ) : (
              (data.incidents ?? []).map((item) => (
                <article key={item.id}>
                  <div>
                    <h3>{item.title}</h3>
                    <span>
                      {item.application.replace('_', ' ')} ·{' '}
                      {item.status === 'resolved' ? 'Resuelto' : 'Investigando'}
                    </span>
                  </div>
                  <p>{item.message}</p>
                  <time>{formatDateTime(item.published_at)}</time>
                </article>
              ))
            )}
          </section>
        </>
      )}
      <footer>Atalaya · Observabilidad operativa</footer>
    </main>
  )
}
