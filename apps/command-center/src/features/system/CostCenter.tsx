import { useEffect, useState } from 'react'
import { api, type CostSummary } from '../../api'
import { applicationNames } from '../../lib/format'

export function CostCenter() {
  const [costs, setCosts] = useState<CostSummary | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    api
      .costs()
      .then(setCosts)
      .catch(() => setError('No se pudieron obtener las métricas de costos de LLM.'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="panel"><p style={{ padding: '1rem' }}>Cargando consumo de LLM...</p></div>
  if (error || !costs) return null

  const usedPercent = Math.min(costs.budget_used_percent, 100)
  const isWarning = costs.budget_used_percent >= 80
  const isCritical = costs.budget_used_percent >= 100

  return (
    <section className="panel" style={{ marginTop: '1.5rem' }}>
      <header className="panel-heading">
        <div>
          <h2>Consumo de LLM (OpenRouter)</h2>
          <small style={{ opacity: 0.8 }}>Monitoreo de tokens, presupuestos y costo acumulado</small>
        </div>
        <span
          className="badge"
          style={{
            background: isCritical ? '#ef444422' : isWarning ? '#f59e0b22' : '#10b98122',
            color: isCritical ? '#ef4444' : isWarning ? '#f59e0b' : '#10b981',
            border: `1px solid ${isCritical ? '#ef444444' : isWarning ? '#f59e0b44' : '#10b98144'}`,
            padding: '0.25rem 0.75rem',
            borderRadius: '4px',
            fontSize: '0.85rem',
          }}
        >
          {costs.budget_used_percent.toFixed(1)}% del presupuesto
        </span>
      </header>

      {/* Monthly Budget Bar */}
      <div style={{ padding: '1rem', borderBottom: '1px solid rgba(255,255,255,0.08)' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem', fontSize: '0.9rem' }}>
          <span>Presupuesto Mensual (${costs.monthly_budget_usd.toFixed(2)} USD)</span>
          <strong>${costs.monthly_cost_usd.toFixed(4)} USD consumidos</strong>
        </div>
        <div style={{ height: '8px', background: 'rgba(255,255,255,0.1)', borderRadius: '4px', overflow: 'hidden' }}>
          <div
            style={{
              height: '100%',
              width: `${usedPercent}%`,
              background: isCritical ? '#ef4444' : isWarning ? '#f59e0b' : '#10b981',
              transition: 'width 0.3s ease',
            }}
          />
        </div>
      </div>

      {/* Metrics Summary Cards */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
          gap: '1rem',
          padding: '1rem',
          borderBottom: '1px solid rgba(255,255,255,0.08)',
        }}
      >
        <div>
          <small style={{ color: 'var(--text-muted)' }}>Costo Total Acumulado</small>
          <div style={{ fontSize: '1.25rem', fontWeight: 600, marginTop: '0.25rem' }}>
            ${costs.total_cost_usd.toFixed(4)} USD
          </div>
        </div>
        <div>
          <small style={{ color: 'var(--text-muted)' }}>Tokens Totales</small>
          <div style={{ fontSize: '1.25rem', fontWeight: 600, marginTop: '0.25rem' }}>
            {costs.total_tokens.toLocaleString()}
          </div>
          <small style={{ fontSize: '0.75rem', opacity: 0.7 }}>
            {costs.input_tokens.toLocaleString()} in / {costs.output_tokens.toLocaleString()} out
          </small>
        </div>
        <div>
          <small style={{ color: 'var(--text-muted)' }}>Interpretaciones</small>
          <div style={{ fontSize: '1.25rem', fontWeight: 600, marginTop: '0.25rem' }}>
            {costs.total_requests} solicitudes
          </div>
        </div>
        <div>
          <small style={{ color: 'var(--text-muted)' }}>Latencia Promedio</small>
          <div style={{ fontSize: '1.25rem', fontWeight: 600, marginTop: '0.25rem' }}>
            {costs.average_latency_ms} ms
          </div>
        </div>
      </div>

      {/* Breakdown tables */}
      <div style={{ padding: '1rem' }}>
        <h3 style={{ fontSize: '1rem', marginBottom: '0.75rem' }}>Desglose por Aplicación</h3>
        {costs.by_application.length === 0 ? (
          <p style={{ opacity: 0.7, fontSize: '0.9rem' }}>Sin consumo registrado por aplicación.</p>
        ) : (
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.9rem' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid rgba(255,255,255,0.1)', textAlign: 'left' }}>
                <th style={{ padding: '0.5rem' }}>Aplicación</th>
                <th style={{ padding: '0.5rem' }}>Tokens</th>
                <th style={{ padding: '0.5rem' }}>Solicitudes</th>
                <th style={{ padding: '0.5rem', textAlign: 'right' }}>Costo Est. (USD)</th>
              </tr>
            </thead>
            <tbody>
              {costs.by_application.map((app) => (
                <tr key={app.application} style={{ borderBottom: '1px solid rgba(255,255,255,0.05)' }}>
                  <td style={{ padding: '0.5rem' }}>
                    <strong>{applicationNames[app.application] || app.application}</strong>
                  </td>
                  <td style={{ padding: '0.5rem' }}>{app.total_tokens.toLocaleString()}</td>
                  <td style={{ padding: '0.5rem' }}>{app.request_count}</td>
                  <td style={{ padding: '0.5rem', textAlign: 'right' }}>${app.estimated_cost_usd.toFixed(4)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        <h3 style={{ fontSize: '1rem', marginTop: '1.5rem', marginBottom: '0.75rem' }}>Desglose por Modelo</h3>
        {costs.by_model.length === 0 ? (
          <p style={{ opacity: 0.7, fontSize: '0.9rem' }}>Sin modelos registrados.</p>
        ) : (
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.9rem' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid rgba(255,255,255,0.1)', textAlign: 'left' }}>
                <th style={{ padding: '0.5rem' }}>Modelo</th>
                <th style={{ padding: '0.5rem' }}>Tokens</th>
                <th style={{ padding: '0.5rem' }}>Solicitudes</th>
                <th style={{ padding: '0.5rem', textAlign: 'right' }}>Costo Est. (USD)</th>
              </tr>
            </thead>
            <tbody>
              {costs.by_model.map((m) => (
                <tr key={m.model} style={{ borderBottom: '1px solid rgba(255,255,255,0.05)' }}>
                  <td style={{ padding: '0.5rem' }}>
                    <code>{m.model}</code>
                  </td>
                  <td style={{ padding: '0.5rem' }}>{m.total_tokens.toLocaleString()}</td>
                  <td style={{ padding: '0.5rem' }}>{m.request_count}</td>
                  <td style={{ padding: '0.5rem', textAlign: 'right' }}>${m.estimated_cost_usd.toFixed(4)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </section>
  )
}
