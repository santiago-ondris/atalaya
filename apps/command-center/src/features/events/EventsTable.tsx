import type { EventSummary } from '../../api'
import { EmptyState } from '../../components/feedback/FeedbackState'
import { applicationNames, formatDateTime, formatStatus } from '../../lib/format'

interface EventsTableProps {
  events: EventSummary[]
  total: number
  offset: number
  limit: number
  onSelectEvent: (id: string) => void
  onPreviousPage: () => void
  onNextPage: () => void
}

export function EventsTable({
  events,
  total,
  offset,
  limit,
  onSelectEvent,
  onPreviousPage,
  onNextPage,
}: EventsTableProps) {
  return (
    <>
      {events.length ? (
        <>
          <div className="event-table header">
            <span>Severidad</span>
            <span>Aplicación</span>
            <span>Evento</span>
            <span>Estado</span>
            <span>Ocurrencias</span>
            <span>Hora</span>
          </div>

          {events.map((event) => (
            <button
              className="event-table"
              onClick={() => onSelectEvent(event.id)}
              key={event.id}
            >
              <span className={`severity severity-${event.severity}`}>
                {event.severity}
              </span>
              <span>{applicationNames[event.application]}</span>
              <strong>{event.message}</strong>
              <span>{formatStatus(event.state)}</span>
              <span className="mono">{event.occurrence_count}</span>
              <time>{formatDateTime(event.occurred_at)}</time>
            </button>
          ))}
        </>
      ) : (
        <EmptyState message="Ningún evento coincide con estos filtros." />
      )}

      <footer className="pagination">
        <span>{formatRange(offset, limit, total)}</span>
        <div>
          <button disabled={offset === 0} onClick={onPreviousPage}>
            Anterior
          </button>
          <button disabled={offset + limit >= total} onClick={onNextPage}>
            Siguiente
          </button>
        </div>
      </footer>
    </>
  )
}

function formatRange(offset: number, limit: number, total: number): string {
  if (!total) return '0 registros'
  return `${offset + 1}–${Math.min(offset + limit, total)} de ${total}`
}
