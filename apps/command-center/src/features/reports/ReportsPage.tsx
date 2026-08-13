import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router'
import { api, type DailyReport } from '../../api'
import { productApplications } from '../../catalog/applications'
import { ErrorState, LoadingState } from '../../components/feedback/FeedbackState'

const statusLabels: Record<DailyReport['status'], string> = {
  collecting: 'Recopilando',
  pending: 'Pendiente',
  processing: 'Enviando',
  sent: 'Enviado',
  expired: 'Vencido',
}

export function ReportsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const application = validApplication(searchParams.get('application'))
  const [reports, setReports] = useState<DailyReport[] | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    api
      .reports()
      .then(({ reports }) => setReports(reports))
      .catch(() => setError('No se pudo cargar el historial de reportes.'))
  }, [])

  function changeApplication(value: string) {
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      if (value) next.set('application', value)
      else next.delete('application')
      return next
    })
  }

  if (error)
    return (
      <main className="page">
        <ErrorState message={error} />
      </main>
    )
  if (!reports) return <LoadingState />

  const visibleReports = application
    ? reports.filter((report) =>
        report.applications.some((item) => item.application === application),
      )
    : reports

  return (
    <main className="page reports-page">
      <header className="page-title">
        <div>
          <span className="eyebrow">BITÁCORA DIARIA</span>
          <h1>Reportes</h1>
        </div>
        <p>Resumen enviado a las 20:00, hora de Argentina.</p>
      </header>
      <label className="report-filter">
        Aplicación
        <select
          aria-label="Aplicación"
          value={application}
          onChange={(event) => changeApplication(event.target.value)}
        >
          <option value="">Todas</option>
          {productApplications.map((app) => (
            <option key={app.slug} value={app.slug}>
              {app.displayName}
            </option>
          ))}
        </select>
      </label>
      {visibleReports.length === 0 ? (
        <section className="report-empty">
          <h2>Todavía no hay reportes{application ? ' para esta aplicación' : ''}</h2>
          <p>El primero se generará a las 20:00 ARG.</p>
        </section>
      ) : (
        visibleReports.map((report) => {
          const applications = application
            ? report.applications.filter((item) => item.application === application)
            : report.applications
          return (
            <article className="report-card" key={report.id}>
              <header>
                <div>
                  <span className="eyebrow">{report.date}</span>
                  <h2>Parte de guardia</h2>
                </div>
                <span className={`report-status ${report.status}`}>
                  {statusLabels[report.status]}
                </span>
              </header>
              <div className="report-app-grid">
                {applications.map((app) => (
                  <section key={app.application}>
                    <h3>{app.display_name}</h3>
                    {app.activity_status === 'available' ? (
                      <>
                        <strong>{app.activity_count ?? '—'}</strong>
                        <span>
                          {' '}
                          {app.activity_kind === 'page_views' ? 'vistas' : 'sesiones'}
                        </span>
                      </>
                    ) : (
                      <p className="unavailable-copy">Actividad no disponible</p>
                    )}
                    <p>
                      {app.error_count} errores · {app.occurrence_count} ocurrencias
                    </p>
                    <small>
                      Críticos {app.severity_counts.critical} · Altos{' '}
                      {app.severity_counts.high} · Accionables {app.actionable_count}
                    </small>
                  </section>
                ))}
              </div>
              <footer>
                {report.sent_at
                  ? `Enviado ${new Date(report.sent_at).toLocaleString('es-AR')}`
                  : `Intentos: ${report.attempts}`}
              </footer>
            </article>
          )
        })
      )}
    </main>
  )
}

function validApplication(value: string | null) {
  return productApplications.some((application) => application.slug === value)
    ? (value as string)
    : ''
}
