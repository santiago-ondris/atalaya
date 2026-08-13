import { useEffect, useState } from 'react'
import { api, type IntegrationStatus, type SystemHealth } from '../../api'
import { ErrorState } from '../../components/feedback/FeedbackState'
import { SignalFlag } from '../../components/status/SignalFlag'
import { applicationNames, formatDateTime, formatStatus } from '../../lib/format'
import { CostCenter } from './CostCenter'

export function SystemPage() {
  const [integrations, setIntegrations] = useState<IntegrationStatus[]>([])
  const [error, setError] = useState('')
  const [health, setHealth] = useState<SystemHealth | null>(null)

  useEffect(() => {
    api
      .integrations()
      .then((response) => setIntegrations(response.integrations))
      .catch(() => setError('No se pudo consultar el estado interno.'))
    api
      .systemHealth()
      .then(setHealth)
      .catch(() => setError('No se pudo consultar el estado interno.'))
  }, [])

  return (
    <main className="page">
      <header className="page-title">
        <div>
          <span className="eyebrow">03 / INSTRUMENTOS</span>
          <h1>Estado del sistema</h1>
          <p>Salud de pollers, fuentes, presupuestos y costos de LLM.</p>
        </div>
      </header>

      <section className="panel">
        {error ? (
          <ErrorState message={error} />
        ) : (
          integrations.map((integration) => (
            <IntegrationRow integration={integration} key={integration.id} />
          ))
        )}
      </section>

      <CostCenter />

      {health && (
        <section className="panel" style={{ marginTop: '1.5rem' }}>
          <header className="panel-heading">
            <h2>Procesos y colas</h2>
            <span>{health.status}</span>
          </header>
          {health.signals.map((signal) => (
            <div className="integration-row" key={signal.name}>
              <SignalFlag status={signal.status} />
              <div>
                <strong>{signal.name.replaceAll('_', ' ')}</strong>
                <small>{signal.detail || 'Sin anomalías'}</small>
              </div>
              <span>{formatStatus(signal.status)}</span>
              <time>{formatDateTime(signal.last_seen_at)}</time>
            </div>
          ))}
          <div className="integration-row">
            <div />
            <div>
              <strong>Colas</strong>
              <small>
                {Object.entries(health.queues)
                  .map(([name, count]) => `${name}: ${count}`)
                  .join(' · ')}
              </small>
            </div>
          </div>
        </section>
      )}
    </main>
  )
}

function IntegrationRow({ integration }: { integration: IntegrationStatus }) {
  return (
    <div className="integration-row">
      <SignalFlag status={integration.status} />
      <div>
        <strong>
          {applicationNames[integration.application]} / {integration.display_name}
        </strong>
        <small>
          {integration.source} · {integration.project}
        </small>
      </div>
      <span>{formatStatus(integration.status)}</span>
      <time>{formatDateTime(integration.last_success_at)}</time>
      {integration.last_error && <p>{integration.last_error}</p>}
    </div>
  )
}
