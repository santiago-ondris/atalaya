import type { Overview } from '../../api'
import { EmptyState } from '../../components/feedback/FeedbackState'
import { SignalFlag } from '../../components/status/SignalFlag'
import { applicationNames, formatDateTime, formatStatus } from '../../lib/format'

interface OverviewPageProps {
  overview: Overview
  onSelectEvent: (id: string) => void
  onSelectArchitecture?: (slug: string) => void
}

export function OverviewPage({
  overview,
  onSelectEvent,
  onSelectArchitecture,
}: OverviewPageProps) {
  const healthyApplications = overview.applications.filter(
    (application) => application.status === 'healthy' || application.status === 'ok',
  ).length

  return (
    <main className="page">
      <header className="page-title">
        <div>
          <span className="eyebrow">01 / ESTADO GENERAL</span>
          <h1>Producción, a la vista</h1>
          <p>Cuatro aplicaciones bajo observación continua.</p>
        </div>
        <span className="timestamp">
          Generado {formatDateTime(overview.generated_at)}
        </span>
      </header>

      <section className="metrics">
        <Metric
          label="Aplicaciones sanas"
          value={`${healthyApplications}/4`}
          description="estado de integraciones"
        />
        <Metric
          label="Eventos registrados"
          value={overview.event_total}
          description="historial completo"
        />
        <Metric
          label="Integraciones"
          value={overview.integrations.length}
          description="Sentry + Azure"
        />
        <Metric
          label="Señales recientes"
          value={overview.recent_events.length}
          description="última consulta"
        />
      </section>

      <div className="overview-grid">
        <ApplicationsPanel
          onSelectArchitecture={onSelectArchitecture}
          overview={overview}
        />
        <RecentEventsPanel overview={overview} onSelectEvent={onSelectEvent} />
      </div>
    </main>
  )
}

interface MetricProps {
  label: string
  value: string | number
  description: string
}

function Metric({ label, value, description }: MetricProps) {
  return (
    <article>
      <span>{label}</span>
      <strong>{value}</strong>
      <small>{description}</small>
    </article>
  )
}

function ApplicationsPanel({
  overview,
  onSelectArchitecture,
}: {
  overview: Overview
  onSelectArchitecture?: (slug: string) => void
}) {
  return (
    <section className="panel">
      <div className="panel-title">
        <h2>Aplicaciones</h2>
        <span>4 horizontes</span>
      </div>

      <div className="app-list">
        {overview.applications.map((application) => (
          <div className="app-row" key={application.slug}>
            <SignalFlag status={application.status} />
            <div>
              <strong>{application.display_name}</strong>
              <small>{formatStatus(application.status)}</small>
            </div>
            <time>{formatDateTime(application.last_success_at)}</time>
            {onSelectArchitecture && (
              <button
                className="app-arch-btn"
                onClick={() => onSelectArchitecture(application.slug)}
                title={`Ver diagrama de arquitectura de ${application.display_name}`}
                type="button"
              >
                <i className="ti ti-sitemap" />
                <span>Arquitectura</span>
              </button>
            )}
          </div>
        ))}
      </div>
    </section>
  )
}

interface RecentEventsPanelProps {
  overview: Overview
  onSelectEvent: (id: string) => void
}

function RecentEventsPanel({ overview, onSelectEvent }: RecentEventsPanelProps) {
  return (
    <section className="panel">
      <div className="panel-title">
        <h2>Últimas señales</h2>
        <span>{overview.event_total} total</span>
      </div>

      {overview.recent_events.length ? (
        <div className="event-list">
          {overview.recent_events.map((event) => (
            <button
              className="event-row"
              onClick={() => onSelectEvent(event.id)}
              key={event.id}
            >
              <span className={`severity severity-${event.severity}`}>
                {event.severity}
              </span>
              <span>
                <strong>{applicationNames[event.application]}</strong>
                {event.message}
              </span>
              <time>{formatDateTime(event.occurred_at)}</time>
            </button>
          ))}
        </div>
      ) : (
        <EmptyState message="Todavía no ingresaron eventos a la bitácora." />
      )}
    </section>
  )
}
