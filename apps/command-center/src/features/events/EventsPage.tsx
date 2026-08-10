import { useEffect, useState } from 'react'
import { api, type EventSummary } from '../../api'
import { ErrorState, LoadingState } from '../../components/feedback/FeedbackState'
import { EventFilters, type EventFilterValues } from './EventFilters'
import { EventsTable } from './EventsTable'

const PAGE_SIZE = 20
const initialFilters: EventFilterValues = {
  application: '',
  severity: '',
  state: '',
  period: '24h',
}

interface EventsPageProps {
  onSelectEvent: (id: string) => void
}

export function EventsPage({ onSelectEvent }: EventsPageProps) {
  const [events, setEvents] = useState<EventSummary[]>([])
  const [total, setTotal] = useState(0)
  const [offset, setOffset] = useState(0)
  const [filters, setFilters] = useState(initialFilters)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    const query = buildQuery(filters, offset)

    setIsLoading(true)
    setError('')

    api
      .events(query)
      .then((response) => {
        setEvents(response.events)
        setTotal(response.total)
      })
      .catch(() => setError('No se pudo consultar la bitácora.'))
      .finally(() => setIsLoading(false))
  }, [filters, offset])

  function handleFilterChange(name: keyof EventFilterValues, value: string) {
    setFilters((current) => ({ ...current, [name]: value }))
    setOffset(0)
  }

  return (
    <main className="page">
      <header className="page-title">
        <div>
          <span className="eyebrow">02 / BITÁCORA</span>
          <h1>Eventos</h1>
          <p>Señales normalizadas, interpretadas y trazables.</p>
        </div>
        <span className="timestamp">{total} registros</span>
      </header>

      <EventFilters values={filters} onChange={handleFilterChange} />

      <section className="panel table-panel">
        {isLoading ? (
          <LoadingState />
        ) : error ? (
          <ErrorState message={error} />
        ) : (
          <EventsTable
            events={events}
            total={total}
            offset={offset}
            limit={PAGE_SIZE}
            onSelectEvent={onSelectEvent}
            onPreviousPage={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
            onNextPage={() => setOffset(offset + PAGE_SIZE)}
          />
        )}
      </section>
    </main>
  )
}

function buildQuery(filters: EventFilterValues, offset: number) {
  const query = new URLSearchParams({
    limit: String(PAGE_SIZE),
    offset: String(offset),
    period: filters.period,
  })

  if (filters.application) query.set('application', filters.application)
  if (filters.severity) query.set('severity', filters.severity)
  if (filters.state) query.set('state', filters.state)

  return query
}
