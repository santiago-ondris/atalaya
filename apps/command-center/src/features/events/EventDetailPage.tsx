import { useEffect, useState } from 'react'
import { api, type EventDetail } from '../../api'
import {
  EmptyState,
  ErrorState,
  LoadingState,
} from '../../components/feedback/FeedbackState'
import { applicationNames, formatDateTime } from '../../lib/format'

interface EventDetailPageProps {
  eventId: string
  onBack: () => void
}

export function EventDetailPage({ eventId, onBack }: EventDetailPageProps) {
  const [event, setEvent] = useState<EventDetail | null>(null)
  const [error, setError] = useState('')
  const [incidentTitle, setIncidentTitle] = useState('')
  const [incidentResult, setIncidentResult] = useState('')

  useEffect(() => {
    api
      .event(eventId)
      .then(setEvent)
      .catch(() => setError('No se pudo cargar el detalle técnico.'))
  }, [eventId])

  if (error) {
    return (
      <main className="page">
        <ErrorState message={error} />
      </main>
    )
  }

  if (!event) {
    return (
      <main className="page">
        <LoadingState />
      </main>
    )
  }

  const currentEvent = event

  const severity = event.interpretation?.severity ?? 'pending'

  async function createIncident() {
    const title =
      incidentTitle.trim() ||
      currentEvent.interpretation?.summary ||
      currentEvent.error_type
    try {
      const incident = await api.createIncident(title, [currentEvent.error_group_id])
      setIncidentResult(`Incidente creado: ${incident.title}`)
    } catch {
      setIncidentResult('Este grupo ya está bajo investigación o no pudo convertirse.')
    }
  }

  return (
    <main className="page">
      <button className="back" onClick={onBack}>
        ← Volver a eventos
      </button>

      <header className="detail-title">
        <div>
          <span className={`severity severity-${severity}`}>{severity}</span>
          <h1>{event.message}</h1>
          <p>
            {applicationNames[event.application]} · {event.component} ·{' '}
            {event.environment}
          </p>
        </div>
        <time>{formatDateTime(event.occurred_at)}</time>
      </header>

      <section className="incident-create panel">
        <div>
          <strong>Abrir investigación</strong>
          <span>Convierte este grupo de error en un incidente.</span>
        </div>
        <input
          value={incidentTitle}
          onChange={(event) => setIncidentTitle(event.target.value)}
          placeholder={event.interpretation?.summary || event.error_type}
          maxLength={160}
        />
        <button onClick={createIncident}>Crear incidente</button>
        {incidentResult && <p>{incidentResult}</p>}
      </section>

      <div className="detail-grid">
        <InterpretationPanel event={event} />
        <TechnicalFacts event={event} />
        <OccurrencesPanel event={event} />
        <StackTracePanel stackTrace={event.stack_trace} />
      </div>
    </main>
  )
}

function InterpretationPanel({ event }: { event: EventDetail }) {
  return (
    <section className="panel interpretation">
      <div className="panel-title">
        <h2>Interpretación</h2>
        <span>{event.interpretation?.model ?? 'Pendiente'}</span>
      </div>

      {event.interpretation ? (
        <>
          <h3>{event.interpretation.summary}</h3>
          <p>{event.interpretation.explanation}</p>
          <h4>Acciones sugeridas</h4>
          <ol>
            {event.interpretation.suggested_actions.map((action) => (
              <li key={action}>{action}</li>
            ))}
          </ol>
        </>
      ) : (
        <EmptyState message="El interpreter todavía no procesó este evento." />
      )}
    </section>
  )
}

function TechnicalFacts({ event }: { event: EventDetail }) {
  return (
    <aside className="panel facts">
      <div className="panel-title">
        <h2>Ficha técnica</h2>
        <span>{event.source_event_id}</span>
      </div>
      <dl>
        <dt>Tipo</dt>
        <dd>{event.error_type}</dd>
        <dt>Release</dt>
        <dd>{event.release || 'Sin release'}</dd>
        <dt>Fingerprint</dt>
        <dd>{event.fingerprint}</dd>
        <dt>Ingresado</dt>
        <dd>{formatDateTime(event.ingested_at)}</dd>
      </dl>
    </aside>
  )
}

function OccurrencesPanel({ event }: { event: EventDetail }) {
  return (
    <section className="panel occurrences">
      <div className="panel-title">
        <h2>Ocurrencias</h2>
        <span>Últimas {event.occurrences.length}</span>
      </div>
      {event.occurrences.map((occurrence) => (
        <div className="occurrence-row" key={occurrence.id}>
          <time>{formatDateTime(occurrence.occurred_at)}</time>
          <span>{occurrence.message}</span>
          <code>{occurrence.source_event_id}</code>
        </div>
      ))}
    </section>
  )
}

function StackTracePanel({ stackTrace }: { stackTrace?: string }) {
  return (
    <section className="panel trace">
      <div className="panel-title">
        <h2>Stack trace</h2>
        <span>Sanitizado</span>
      </div>
      <pre>{stackTrace || 'No se recibió stack trace para este evento.'}</pre>
    </section>
  )
}
