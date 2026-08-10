import { useEffect, useState } from 'react'
import { api, type IntegrationStatus } from '../../api'
import { ErrorState } from '../../components/feedback/FeedbackState'
import { SignalFlag } from '../../components/status/SignalFlag'
import { applicationNames, formatDateTime, formatStatus } from '../../lib/format'

export function SystemPage() {
  const [integrations, setIntegrations] = useState<IntegrationStatus[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    api
      .integrations()
      .then((response) => setIntegrations(response.integrations))
      .catch(() => setError('No se pudo consultar el estado interno.'))
  }, [])

  return (
    <main className="page">
      <header className="page-title">
        <div>
          <span className="eyebrow">03 / INSTRUMENTOS</span>
          <h1>Estado del sistema</h1>
          <p>Salud de pollers, fuentes y checkpoints.</p>
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
