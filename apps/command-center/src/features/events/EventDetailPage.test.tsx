import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { api, type EventDetail, type Incident } from '../../api'
import { EventDetailPage } from './EventDetailPage'

const event: EventDetail = {
  id: 'event-1',
  error_group_id: 'group-1',
  source_event_id: 'source-1',
  application: 'prensap',
  component: 'frontend',
  environment: 'production',
  error_type: 'TypeError',
  message: 'Falló la portada',
  occurred_at: '2026-08-11T12:00:00Z',
  ingested_at: '2026-08-11T12:01:00Z',
  severity: 'high',
  state: 'actionable',
  occurrence_count: 3,
  fingerprint: 'fingerprint',
  metadata: {},
  occurrences: [],
  interpretation: {
    summary: 'La portada no carga',
    explanation: 'Un acceso inválido.',
    severity: 'high',
    actionable: true,
    suggested_actions: ['Revisar datos'],
    model: 'test',
    total_tokens: 10,
    latency_ms: 20,
  },
}

afterEach(() => vi.restoreAllMocks())

describe('EventDetailPage', () => {
  it('creates an incident from the underlying error group', async () => {
    vi.spyOn(api, 'event').mockResolvedValue(event)
    vi.spyOn(api, 'createIncident').mockResolvedValue({
      id: 'incident-1',
      application: 'prensap',
      title: 'La portada no carga',
      status: 'investigating',
      created_at: event.occurred_at,
      updated_at: event.occurred_at,
    } satisfies Incident)
    render(<EventDetailPage eventId="event-1" onBack={() => undefined} />)
    await screen.findByText('Falló la portada')
    fireEvent.click(screen.getByRole('button', { name: 'Crear incidente' }))
    await waitFor(() =>
      expect(api.createIncident).toHaveBeenCalledWith('La portada no carga', ['group-1']),
    )
    expect(
      await screen.findByText('Incidente creado: La portada no carga'),
    ).toBeInTheDocument()
  })
})
