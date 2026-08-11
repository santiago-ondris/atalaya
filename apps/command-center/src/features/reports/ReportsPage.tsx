import { useEffect, useState } from 'react'
import { api, type DailyReport } from '../../api'
import { ErrorState, LoadingState } from '../../components/feedback/FeedbackState'

const statusLabels: Record<DailyReport['status'], string> = {
  collecting: 'Recopilando',
  pending: 'Pendiente',
  processing: 'Enviando',
  sent: 'Enviado',
  expired: 'Vencido',
}

export function ReportsPage() {
  const [reports, setReports] = useState<DailyReport[] | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    api
      .reports()
      .then(({ reports }) => setReports(reports))
      .catch(() => setError('No se pudo cargar el historial de reportes.'))
  }, [])

  if (error)
    return (
      <main className="page">
        <ErrorState message={error} />
      </main>
    )
  if (!reports) return <LoadingState />

  return (
    <main className="page reports-page">
      <header className="page-title">
        <div>
          <span className="eyebrow">BITÁCORA DIARIA</span>
          <h1>Reportes</h1>
        </div>
        <p>Resumen enviado a las 20:00, hora de Argentina.</p>
      </header>
      {reports.length === 0 ? (
        <section className="report-empty">
          <h2>Todavía no hay reportes</h2>
          <p>El primero se generará a las 20:00 ARG.</p>
        </section>
      ) : (
        reports.map((report) => (
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
              {report.applications.map((app) => (
                <section key={app.application}>
                  <h3>{app.display_name}</h3>
                  <strong>{app.activity_count ?? '—'}</strong>
                  <span> sesiones</span>
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
        ))
      )}
    </main>
  )
}
