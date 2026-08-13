import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { api, type DailyReport, type PublicStatus } from '../../api'
import { resolveApplication } from '../../catalog/applications'
import { ApplicationPage } from './ApplicationPage'

const publicStatus: PublicStatus = {
  status: 'operational',
  generated_at: '2026-08-13T12:00:00Z',
  applications: [
    {
      slug: 'farmami',
      display_name: 'Farmami',
      status: 'operational',
      uptime_30_days: null,
      last_checked_at: '2026-08-13T12:00:00Z',
      components: [
        { name: 'frontend', status: 'up' },
        { name: 'backend', status: 'down' },
      ],
    },
  ],
  incidents: [],
}

const report: DailyReport = {
  id: 'report-1',
  date: '2026-08-12',
  timezone: 'America/Argentina/Cordoba',
  period_start: '2026-08-12T00:00:00Z',
  period_end: '2026-08-13T00:00:00Z',
  status: 'sent',
  attempts: 1,
  created_at: '2026-08-13T00:00:00Z',
  applications: [
    {
      application: 'farmami',
      display_name: 'Farmami',
      activity_kind: 'sessions',
      activity_source: 'analytics',
      activity_status: 'unavailable',
      activity_error: 'timeout',
      error_count: 2,
      occurrence_count: 5,
      severity_counts: { critical: 1, high: 1, medium: 0, low: 0 },
      actionable_count: 2,
    },
  ],
}

beforeEach(() => {
  vi.spyOn(api, 'publicStatus').mockResolvedValue(publicStatus)
  vi.spyOn(api, 'integrations').mockResolvedValue({
    integrations: [
      {
        id: 'integration-1',
        application: 'farmami',
        component: 'frontend',
        display_name: 'Farmami frontend',
        source: 'sentry',
        project: 'farmami',
        enabled: true,
        environments: ['production'],
        status: 'error',
        last_error: 'invalid token',
      },
    ],
  })
  vi.spyOn(api, 'reports').mockResolvedValue({ reports: [report] })
})

afterEach(() => {
  cleanup()
  vi.useRealTimers()
  vi.restoreAllMocks()
})

describe('ApplicationPage', () => {
  it('renders product operations and explicit incomplete states', async () => {
    render(
      <MemoryRouter>
        <ApplicationPage app={resolveApplication('farmami')!} />
      </MemoryRouter>,
    )

    expect(await screen.findByText('Disponibilidad · 30 días')).toBeInTheDocument()
    expect(screen.getAllByText('Sin datos').length).toBeGreaterThan(0)
    expect(screen.getByText(/Error operativo: invalid token/)).toBeInTheDocument()
    expect(screen.getByText(/Actividad no disponible: timeout/)).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Eventos/ })).toHaveAttribute(
      'href',
      '/events?application=farmami',
    )
  })

  it('keeps a failed panel isolated', async () => {
    vi.mocked(api.integrations).mockRejectedValue(new Error('offline'))
    render(
      <MemoryRouter>
        <ApplicationPage app={resolveApplication('farmami')!} />
      </MemoryRouter>,
    )
    expect(
      await screen.findByText('No se pudieron consultar las integraciones.'),
    ).toBeInTheDocument()
    expect(screen.getByText('Disponibilidad · 30 días')).toBeInTheDocument()
  })

  it('polls live sources and cancels the timer on unmount', async () => {
    vi.useFakeTimers()
    const clear = vi.spyOn(window, 'clearInterval')
    const view = render(
      <MemoryRouter>
        <ApplicationPage app={resolveApplication('farmami')!} />
      </MemoryRouter>,
    )
    await vi.advanceTimersByTimeAsync(60_000)
    expect(api.publicStatus).toHaveBeenCalledTimes(2)
    expect(api.integrations).toHaveBeenCalledTimes(2)
    view.unmount()
    expect(clear).toHaveBeenCalled()
  })

  it('renders platform health, queues and costs without product reports', async () => {
    vi.spyOn(api, 'systemHealth').mockResolvedValue({
      status: 'healthy',
      generated_at: '2026-08-13T12:00:00Z',
      signals: [{ name: 'watchman', status: 'healthy', detail: 'online' }],
      queues: { interpretation: 3 },
    })
    vi.spyOn(api, 'costs').mockResolvedValue({
      total_cost_usd: 8,
      monthly_cost_usd: 4,
      monthly_budget_usd: 20,
      budget_used_percent: 20,
      total_tokens: 100,
      input_tokens: 60,
      output_tokens: 40,
      total_requests: 10,
      average_latency_ms: 125,
      by_application: [],
      by_model: [],
    })
    render(
      <MemoryRouter>
        <ApplicationPage app={resolveApplication('atalaya')!} />
      </MemoryRouter>,
    )
    expect(await screen.findByText('watchman')).toBeInTheDocument()
    expect(screen.getByText('interpretation')).toBeInTheDocument()
    expect(screen.getByText(/no genera reportes diarios de producto/)).toBeInTheDocument()
    await waitFor(() => expect(screen.getByText('20.0%')).toBeInTheDocument())
    expect(api.reports).not.toHaveBeenCalled()
  })
})
