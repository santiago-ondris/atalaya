import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router'
import {
  Bar,
  CartesianGrid,
  ComposedChart,
  Line,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import {
  api,
  type CostSummary,
  type DailyReport,
  type DailyReportApplication,
  type IntegrationStatus,
  type PublicStatus,
  type SystemHealth,
} from '../../api'
import type { Application } from '../../catalog/applications'
import { ErrorState, LoadingState } from '../../components/feedback/FeedbackState'
import { formatDateTime, formatStatus } from '../../lib/format'

const POLL_INTERVAL = 60_000

type Loadable<T> = { data: T | null; loading: boolean; error: string }
const initialLoadable = <T,>(): Loadable<T> => ({
  data: null,
  loading: true,
  error: '',
})

export function ApplicationPage({ app }: { app: Application }) {
  return app.kind === 'platform' ? (
    <PlatformApplicationPage app={app} />
  ) : (
    <ProductApplicationPage app={app} />
  )
}

function ProductApplicationPage({ app }: { app: Application }) {
  const [availability, setAvailability] = useState(initialLoadable<PublicStatus>)
  const [integrations, setIntegrations] = useState(initialLoadable<IntegrationStatus[]>)
  const [reports, setReports] = useState(initialLoadable<DailyReport[]>)

  const loadAvailability = useCallback(() => {
    api.publicStatus().then(
      (data) => setAvailability({ data, loading: false, error: '' }),
      () =>
        setAvailability((current) => ({
          ...current,
          loading: false,
          error: 'No se pudo consultar la disponibilidad.',
        })),
    )
  }, [])
  const loadIntegrations = useCallback(() => {
    api.integrations().then(
      ({ integrations: data }) => setIntegrations({ data, loading: false, error: '' }),
      () =>
        setIntegrations((current) => ({
          ...current,
          loading: false,
          error: 'No se pudieron consultar las integraciones.',
        })),
    )
  }, [])

  useEffect(() => {
    loadAvailability()
    loadIntegrations()
    api.reports().then(
      ({ reports: data }) => setReports({ data, loading: false, error: '' }),
      () =>
        setReports({
          data: null,
          loading: false,
          error: 'No se pudieron cargar los reportes diarios.',
        }),
    )
    const timer = window.setInterval(() => {
      loadAvailability()
      loadIntegrations()
    }, POLL_INTERVAL)
    return () => window.clearInterval(timer)
  }, [loadAvailability, loadIntegrations])

  const status = availability.data?.applications.find((item) => item.slug === app.slug)
  const appIntegrations =
    integrations.data?.filter((item) => item.application === app.slug) ?? []
  const reportEntries = useMemo(
    () => reportHistory(reports.data ?? [], app.slug),
    [app.slug, reports.data],
  )
  const latest = reportEntries.at(-1)

  return (
    <main className="page application-page">
      <ApplicationHeader app={app}>
        <StatusBadge label="Disponibilidad" value={status?.status ?? 'unknown'} />
        <StatusBadge
          label="Observabilidad"
          value={integrationSummary(appIntegrations, integrations)}
        />
      </ApplicationHeader>

      <ContextLinks app={app} platform={false} />

      <div className="application-grid">
        <Panel title="Disponibilidad · 30 días" state={availability}>
          {!status ? (
            <EmptyMessage>
              No hay datos de disponibilidad para esta aplicación.
            </EmptyMessage>
          ) : (
            <>
              <div className="operational-metrics">
                <Metric
                  label="Uptime"
                  value={
                    status.uptime_30_days == null
                      ? 'Sin datos'
                      : `${status.uptime_30_days.toFixed(2)}%`
                  }
                />
                {status.components.map((component) => (
                  <Metric
                    key={component.name}
                    label={component.name}
                    value={
                      component.status === 'up'
                        ? 'Operativo'
                        : formatStatus(component.status)
                    }
                  />
                ))}
              </div>
              <p className="panel-note">
                Última comprobación: {formatDateTime(status.last_checked_at)}
              </p>
            </>
          )}
        </Panel>

        <Panel title="Integraciones" state={integrations}>
          {appIntegrations.length === 0 ? (
            <EmptyMessage>No hay integraciones configuradas.</EmptyMessage>
          ) : (
            <div className="integration-list">
              {appIntegrations.map((integration) => (
                <article key={integration.id}>
                  <header>
                    <strong>{integration.display_name}</strong>
                    <span
                      className={`signal-pill ${integration.enabled ? integration.status : 'disabled'}`}
                    >
                      {integration.enabled
                        ? formatStatus(integration.status)
                        : 'Deshabilitada'}
                    </span>
                  </header>
                  <p>
                    {integration.component} · {integration.source}
                  </p>
                  <dl>
                    <div>
                      <dt>Último éxito</dt>
                      <dd>{formatDateTime(integration.last_success_at)}</dd>
                    </div>
                    <div>
                      <dt>Entornos</dt>
                      <dd>{integration.environments.join(', ') || 'Sin datos'}</dd>
                    </div>
                  </dl>
                  {integration.last_error && (
                    <p className="operation-error">
                      Error operativo: {integration.last_error}
                    </p>
                  )}
                </article>
              ))}
            </div>
          )}
        </Panel>
      </div>

      <Panel title="Último parte diario" state={reports}>
        {!latest ? (
          <EmptyMessage>Esta aplicación todavía no tiene reportes diarios.</EmptyMessage>
        ) : (
          <ReportSummary date={latest.date} report={latest.application} />
        )}
      </Panel>

      <Panel title="Tendencia · últimos siete reportes" state={reports}>
        {reportEntries.length === 0 ? (
          <EmptyMessage>
            No hay reportes suficientes para mostrar una tendencia.
          </EmptyMessage>
        ) : (
          <TrendChart entries={reportEntries} />
        )}
      </Panel>
    </main>
  )
}

function PlatformApplicationPage({ app }: { app: Application }) {
  const [health, setHealth] = useState(initialLoadable<SystemHealth>)
  const [costs, setCosts] = useState(initialLoadable<CostSummary>)

  useEffect(() => {
    api.systemHealth().then(
      (data) => setHealth({ data, loading: false, error: '' }),
      () =>
        setHealth({
          data: null,
          loading: false,
          error: 'No se pudo cargar la salud interna.',
        }),
    )
    api.costs().then(
      (data) => setCosts({ data, loading: false, error: '' }),
      () =>
        setCosts({
          data: null,
          loading: false,
          error: 'No se pudieron cargar los costos.',
        }),
    )
  }, [])

  return (
    <main className="page application-page">
      <ApplicationHeader app={app}>
        <StatusBadge label="Salud interna" value={health.data?.status ?? 'unknown'} />
      </ApplicationHeader>
      <ContextLinks app={app} platform />
      <div className="application-grid">
        <Panel title="Salud interna y señales" state={health}>
          {health.data && (
            <div className="signal-list">
              {health.data.signals.length === 0 ? (
                <EmptyMessage>No hay señales internas disponibles.</EmptyMessage>
              ) : (
                health.data.signals.map((signal) => (
                  <article key={signal.name}>
                    <div>
                      <strong>{signal.name}</strong>
                      <p>{signal.detail || 'Sin detalle'}</p>
                    </div>
                    <span className={`signal-pill ${signal.status}`}>
                      {formatStatus(signal.status)}
                    </span>
                  </article>
                ))
              )}
            </div>
          )}
        </Panel>
        <Panel title="Colas" state={health}>
          {health.data && (
            <div className="operational-metrics">
              {Object.keys(health.data.queues).length === 0 ? (
                <EmptyMessage>No hay métricas de colas disponibles.</EmptyMessage>
              ) : (
                Object.entries(health.data.queues).map(([name, count]) => (
                  <Metric key={name} label={name} value={String(count)} />
                ))
              )}
            </div>
          )}
        </Panel>
      </div>
      <Panel title="Consumo mensual" state={costs}>
        {costs.data && (
          <div className="operational-metrics cost-metrics">
            <Metric
              label="Costo mensual"
              value={formatCurrency(costs.data.monthly_cost_usd)}
            />
            <Metric
              label="Presupuesto utilizado"
              value={`${costs.data.budget_used_percent.toFixed(1)}%`}
            />
            <Metric
              label="Requests"
              value={costs.data.total_requests.toLocaleString('es-AR')}
            />
            <Metric
              label="Latencia promedio"
              value={`${Math.round(costs.data.average_latency_ms)} ms`}
            />
          </div>
        )}
      </Panel>
      <aside className="platform-note">
        Atalaya es la plataforma de observabilidad y no genera reportes diarios de
        producto.
      </aside>
    </main>
  )
}

function ApplicationHeader({
  app,
  children,
}: {
  app: Application
  children: React.ReactNode
}) {
  return (
    <header
      className="application-header"
      style={{ '--app-color': app.brandColor } as React.CSSProperties}
    >
      <div>
        <span className="eyebrow">APLICACIÓN / {app.slug.toUpperCase()}</span>
        <h1>{app.displayName}</h1>
        <p>
          {app.stack} · {app.deploy}
        </p>
      </div>
      <div className="application-statuses">{children}</div>
    </header>
  )
}

function ContextLinks({ app, platform }: { app: Application; platform: boolean }) {
  const links = platform
    ? [
        ['Sistema', '/system'],
        ['Arquitectura', `/architecture/${app.slug}`],
      ]
    : [
        ['Eventos', `/events?application=${app.slug}`],
        ['Bitácora', `/operations?application=${app.slug}`],
        ['Reportes', `/reports?application=${app.slug}`],
        ['Arquitectura', `/architecture/${app.slug}`],
      ]
  return (
    <nav className="context-links" aria-label={`Accesos de ${app.displayName}`}>
      {links.map(([label, path]) => (
        <Link key={label} to={path}>
          {label}
          <span aria-hidden="true">↗</span>
        </Link>
      ))}
    </nav>
  )
}

function Panel<T>({
  title,
  state,
  children,
}: {
  title: string
  state: Loadable<T>
  children: React.ReactNode
}) {
  return (
    <section className="panel operational-panel">
      <div className="panel-title">
        <h2>{title}</h2>
      </div>
      {state.loading ? (
        <LoadingState />
      ) : state.error ? (
        <ErrorState message={state.error} />
      ) : (
        children
      )}
    </section>
  )
}

function StatusBadge({ label, value }: { label: string; value: string }) {
  return (
    <div className="status-badge">
      <span>{label}</span>
      <strong className={value}>{formatStatus(value)}</strong>
    </div>
  )
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <article className="operational-metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </article>
  )
}

function EmptyMessage({ children }: { children: React.ReactNode }) {
  return <p className="empty-message">{children}</p>
}

function ReportSummary({
  date,
  report,
}: {
  date: string
  report: DailyReportApplication
}) {
  const activityLabel = report.activity_kind === 'page_views' ? 'Vistas' : 'Sesiones'
  return (
    <div className="daily-summary">
      <Metric
        label="Fecha"
        value={new Intl.DateTimeFormat('es-AR', { dateStyle: 'medium' }).format(
          new Date(`${date}T12:00:00`),
        )}
      />
      <Metric
        label={activityLabel}
        value={
          report.activity_status === 'available' && report.activity_count != null
            ? String(report.activity_count)
            : 'Sin datos'
        }
      />
      <Metric label="Errores" value={String(report.error_count)} />
      <Metric label="Ocurrencias" value={String(report.occurrence_count)} />
      <Metric
        label="Críticos / altos"
        value={`${report.severity_counts.critical} / ${report.severity_counts.high}`}
      />
      <Metric label="Accionables" value={String(report.actionable_count)} />
      {report.activity_status === 'unavailable' && (
        <p className="operation-error">
          Actividad no disponible
          {report.activity_error ? `: ${report.activity_error}` : '.'}
        </p>
      )}
    </div>
  )
}

function TrendChart({ entries }: { entries: ReportEntry[] }) {
  const data = entries.map((entry) => ({
    date: entry.date,
    activity:
      entry.application.activity_status === 'available'
        ? entry.application.activity_count
        : undefined,
    errors: entry.application.error_count,
  }))
  return (
    <>
      <div
        className="trend-chart"
        role="img"
        aria-label="Tendencia diaria de actividad y errores"
      >
        <ResponsiveContainer width="100%" height="100%">
          <ComposedChart
            data={data}
            accessibilityLayer
            margin={{ top: 16, right: 8, bottom: 4, left: 0 }}
          >
            <CartesianGrid strokeDasharray="3 3" vertical={false} />
            <XAxis dataKey="date" tickFormatter={shortDate} />
            <YAxis yAxisId="activity" width={42} allowDecimals={false} />
            <YAxis
              yAxisId="errors"
              orientation="right"
              width={36}
              allowDecimals={false}
            />
            <Tooltip labelFormatter={(value) => String(value)} />
            <Bar
              yAxisId="errors"
              dataKey="errors"
              name="Errores"
              fill="#C1432E"
              radius={[2, 2, 0, 0]}
            />
            <Line
              yAxisId="activity"
              dataKey="activity"
              name="Actividad"
              stroke="#4750A8"
              strokeWidth={3}
              connectNulls={false}
            />
          </ComposedChart>
        </ResponsiveContainer>
      </div>
      <details className="accessible-data">
        <summary>Ver resumen accesible de la tendencia</summary>
        <table>
          <thead>
            <tr>
              <th>Fecha</th>
              <th>Actividad</th>
              <th>Errores</th>
            </tr>
          </thead>
          <tbody>
            {data.map((item) => (
              <tr key={item.date}>
                <td>{item.date}</td>
                <td>{item.activity ?? 'Sin datos'}</td>
                <td>{item.errors}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </details>
    </>
  )
}

interface ReportEntry {
  date: string
  application: DailyReportApplication
}
function reportHistory(reports: DailyReport[], slug: string): ReportEntry[] {
  return reports
    .flatMap((report) => {
      const application = report.applications.find((item) => item.application === slug)
      return application ? [{ date: report.date, application }] : []
    })
    .sort((a, b) => a.date.localeCompare(b.date))
    .slice(-7)
}

function integrationSummary(
  items: IntegrationStatus[],
  state: Loadable<IntegrationStatus[]>,
) {
  if (state.loading || state.error || !state.data) return 'unknown'
  if (items.length === 0) return 'unconfigured'
  if (items.some((item) => item.enabled && !['ok', 'healthy'].includes(item.status)))
    return 'degraded'
  if (items.every((item) => !item.enabled)) return 'unconfigured'
  return 'healthy'
}

function shortDate(value: string) {
  return new Intl.DateTimeFormat('es-AR', { day: '2-digit', month: 'short' }).format(
    new Date(`${value}T12:00:00`),
  )
}
function formatCurrency(value: number) {
  return new Intl.NumberFormat('es-AR', { style: 'currency', currency: 'USD' }).format(
    value,
  )
}
