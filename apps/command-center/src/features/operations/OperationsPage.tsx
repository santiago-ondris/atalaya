import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Bar,
  BarChart,
  CartesianGrid,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import {
  api,
  type Deployment,
  type Incident,
  type IncidentStatus,
  type OperationsTimeline,
} from '../../api'
import { ErrorState, LoadingState } from '../../components/feedback/FeedbackState'
import { applicationNames, formatDateTime } from '../../lib/format'

const apps = ['farmami', 'wheels_house', 'prensap', 'notizap']

export function OperationsPage() {
  const [application, setApplication] = useState('prensap')
  const [component, setComponent] = useState('')
  const [period, setPeriod] = useState('168h')
  const [data, setData] = useState<OperationsTimeline | null>(null)
  const [selected, setSelected] = useState<Incident | null>(null)
  const [error, setError] = useState('')
  const [deployOpen, setDeployOpen] = useState(false)

  const load = useCallback(() => {
    setError('')
    api
      .timeline(application, component, period)
      .then(setData)
      .catch(() => setError('No se pudo cargar la bitácora.'))
  }, [application, component, period])
  useEffect(load, [load])

  const markers = useMemo(
    () =>
      data?.timeline.deployments.map((deployment) => {
        const deployed = new Date(deployment.deployed_at).getTime()
        const bucket = data.timeline.buckets.findLast(
          (item) => new Date(item.start).getTime() <= deployed,
        )
        return { ...deployment, bucket: bucket?.start }
      }) ?? [],
    [data],
  )

  if (error)
    return (
      <main className="page">
        <ErrorState message={error} />
      </main>
    )
  if (!data)
    return (
      <main className="page">
        <LoadingState />
      </main>
    )

  return (
    <main className="page operations-page">
      <header className="page-title">
        <div>
          <span className="eyebrow">Memoria operativa</span>
          <h1>Bitácora</h1>
          <p>Errores, cambios de producción e investigación en una misma carta.</p>
        </div>
        <button className="primary-action" onClick={() => setDeployOpen(true)}>
          Registrar deploy
        </button>
      </header>
      <section className="operations-filters" aria-label="Filtros de bitácora">
        <label>
          Aplicación
          <select
            value={application}
            onChange={(event) => {
              setApplication(event.target.value)
              setSelected(null)
            }}
          >
            {apps.map((app) => (
              <option key={app} value={app}>
                {applicationNames[app]}
              </option>
            ))}
          </select>
        </label>
        <label>
          Componente
          <select
            value={component}
            onChange={(event) => setComponent(event.target.value)}
          >
            <option value="">Todos</option>
            <option value="frontend">Frontend</option>
            <option value="backend">Backend</option>
          </select>
        </label>
        <label>
          Período
          <select value={period} onChange={(event) => setPeriod(event.target.value)}>
            <option value="24h">24 horas</option>
            <option value="168h">7 días</option>
            <option value="720h">30 días</option>
          </select>
        </label>
      </section>

      <section className="panel operations-chart">
        <div className="panel-title">
          <h2>Ocurrencias y deploys</h2>
          <span>{data.timeline.deployments.length} markers</span>
        </div>
        <div
          className="chart-canvas"
          role="img"
          aria-label="Gráfico temporal de errores y deploys"
        >
          <ResponsiveContainer width="100%" height="100%">
            <BarChart
              data={data.timeline.buckets}
              margin={{ top: 18, right: 16, left: 0, bottom: 6 }}
              accessibilityLayer
            >
              <CartesianGrid strokeDasharray="3 3" vertical={false} />
              <XAxis
                dataKey="start"
                tickFormatter={(value: string) =>
                  new Intl.DateTimeFormat('es-AR', {
                    month: 'short',
                    day: '2-digit',
                    hour: '2-digit',
                  }).format(new Date(value))
                }
                minTickGap={32}
              />
              <YAxis allowDecimals={false} width={32} />
              <Tooltip
                labelFormatter={(value: unknown) => formatDateTime(String(value))}
              />
              <Bar
                dataKey="error_count"
                name="Errores"
                fill="#C1432E"
                radius={[2, 2, 0, 0]}
              />
              {markers
                .filter((marker) => marker.bucket)
                .map((marker) => (
                  <ReferenceLine
                    key={marker.id}
                    x={marker.bucket}
                    stroke="#4750A8"
                    strokeWidth={2}
                    label={{ value: '▲', fill: '#4750A8', position: 'top' }}
                  />
                ))}
            </BarChart>
          </ResponsiveContainer>
        </div>
        <div className="deploy-strip">
          {data.timeline.deployments.length === 0 ? (
            <p>Sin deploys registrados en este período.</p>
          ) : (
            data.timeline.deployments.map((deployment) => (
              <DeploymentMarker key={deployment.id} deployment={deployment} />
            ))
          )}
        </div>
        <details className="accessible-data">
          <summary>Ver datos del gráfico como tabla</summary>
          <table>
            <thead>
              <tr>
                <th>Período</th>
                <th>Errores</th>
              </tr>
            </thead>
            <tbody>
              {data.timeline.buckets.map((bucket) => (
                <tr key={bucket.start}>
                  <td>{formatDateTime(bucket.start)}</td>
                  <td>{bucket.error_count}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </details>
      </section>

      <section className="incident-section">
        <div className="panel-title">
          <h2>Incidentes</h2>
          <span>{data.timeline.incidents.length} entradas</span>
        </div>
        {data.timeline.incidents.length === 0 ? (
          <div className="panel empty-operation">
            Todavía no hay incidentes para {applicationNames[application]}.
          </div>
        ) : (
          <div className="incident-list">
            {data.timeline.incidents.map((incident) => (
              <button
                key={incident.id}
                onClick={() => api.incident(incident.id).then(setSelected)}
                className={selected?.id === incident.id ? 'selected' : ''}
              >
                <span className={`incident-status ${incident.status}`}>
                  {statusLabel(incident.status)}
                </span>
                <strong>{incident.title}</strong>
                <time>{formatDateTime(incident.created_at)}</time>
              </button>
            ))}
          </div>
        )}
      </section>
      {selected && (
        <IncidentPanel
          incident={selected}
          onChanged={() =>
            api.incident(selected.id).then((item) => {
              setSelected(item)
              load()
            })
          }
        />
      )}
      {deployOpen && (
        <DeploymentDialog
          application={application}
          onClose={() => setDeployOpen(false)}
          onCreated={() => {
            setDeployOpen(false)
            load()
          }}
        />
      )}
    </main>
  )
}

const providerNames: Record<Deployment['provider'], string> = {
  railway: 'Railway',
  github_actions: 'GitHub Actions',
  manual: 'Manual',
}

export function DeploymentMarker({ deployment }: { deployment: Deployment }) {
  const shortCommit = deployment.commit_sha?.slice(0, 8)
  const reference = deployment.version || shortCommit || deployment.external_id

  return (
    <article className="deployment-marker">
      <div className="deployment-marker-heading">
        <strong>{reference}</strong>
        <span className="deployment-environment">{deployment.environment}</span>
      </div>
      <dl>
        <div>
          <dt>Componente</dt>
          <dd>{deployment.component}</dd>
        </div>
        <div>
          <dt>Proveedor</dt>
          <dd>{providerNames[deployment.provider]}</dd>
        </div>
        {deployment.actor && (
          <div>
            <dt>Actor</dt>
            <dd>{deployment.actor}</dd>
          </div>
        )}
      </dl>
      <time dateTime={deployment.deployed_at}>
        {formatDateTime(deployment.deployed_at)}
      </time>
      {(deployment.commit_url || deployment.source_url) && (
        <div className="deployment-links">
          {deployment.commit_url && (
            <a href={deployment.commit_url} target="_blank" rel="noreferrer">
              Commit{shortCommit ? ` ${shortCommit}` : ''}
            </a>
          )}
          {deployment.source_url && (
            <a href={deployment.source_url} target="_blank" rel="noreferrer">
              Abrir deployment
            </a>
          )}
        </div>
      )}
    </article>
  )
}

function statusLabel(status: IncidentStatus) {
  return status === 'investigating'
    ? 'Investigando'
    : status === 'resolved'
      ? 'Resuelto'
      : 'Ruido'
}

function IncidentPanel({
  incident,
  onChanged,
}: {
  incident: Incident
  onChanged: () => void
}) {
  const [note, setNote] = useState('')
  const [query, setQuery] = useState('')
  const [groups, setGroups] = useState<
    Awaited<ReturnType<typeof api.errorGroups>>['groups']
  >([])
  const [error, setError] = useState('')
  async function transition(status: IncidentStatus) {
    if (!note.trim()) {
      setError('La conclusión o motivo es obligatorio.')
      return
    }
    try {
      await api.changeIncidentStatus(incident.id, status, note)
      setNote('')
      setError('')
      onChanged()
    } catch {
      setError(
        'No se pudo cambiar el estado; revisá si un grupo ya está bajo investigación.',
      )
    }
  }
  async function search() {
    const result = await api.errorGroups(incident.application, query)
    setGroups(
      result.groups.filter(
        (group) => !incident.groups?.some((item) => item.id === group.id),
      ),
    )
  }
  return (
    <section className="panel incident-detail">
      <div className="panel-title">
        <div>
          <span className={`incident-status ${incident.status}`}>
            {statusLabel(incident.status)}
          </span>
          <h2>{incident.title}</h2>
        </div>
        <span>{incident.groups?.length ?? 0} grupos</span>
      </div>
      <div className="incident-groups">
        {incident.groups?.map((group) => (
          <article key={group.id}>
            <div>
              <strong>{group.error_type}</strong>
              <p>{group.message}</p>
            </div>
            {incident.status === 'investigating' &&
              (incident.groups?.length ?? 0) > 1 && (
                <button
                  onClick={() =>
                    api.removeIncidentGroup(incident.id, group.id).then(onChanged)
                  }
                >
                  Quitar
                </button>
              )}
          </article>
        ))}
      </div>
      {incident.status === 'investigating' && (
        <div className="group-search">
          <input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Buscar otro grupo de error"
          />
          <button onClick={search}>Buscar</button>
          {groups.map((group) => (
            <button
              className="search-result"
              key={group.id}
              onClick={() => api.addIncidentGroup(incident.id, group.id).then(onChanged)}
            >
              <strong>{group.error_type}</strong>
              <span>{group.message}</span>
            </button>
          ))}
        </div>
      )}
      <div className="incident-compose">
        <textarea
          value={note}
          onChange={(event) => setNote(event.target.value)}
          placeholder={
            incident.status === 'investigating'
              ? 'Nota o conclusión de la investigación'
              : 'Motivo para reabrir'
          }
        />
        {error && <p className="form-error">{error}</p>}
        <div>
          <button
            onClick={() =>
              api.addIncidentNote(incident.id, note).then(() => {
                setNote('')
                onChanged()
              })
            }
          >
            Agregar nota
          </button>
          {incident.status === 'investigating' ? (
            <>
              <button onClick={() => transition('resolved')}>Resolver</button>
              <button onClick={() => transition('noise')}>Marcar ruido</button>
            </>
          ) : (
            <button onClick={() => transition('investigating')}>Reabrir</button>
          )}
        </div>
      </div>
      <ol className="incident-log">
        {incident.entries
          ?.slice()
          .reverse()
          .map((entry) => (
            <li key={entry.id}>
              <time>{formatDateTime(entry.created_at)}</time>
              <strong>
                {entry.kind === 'status_change'
                  ? `${statusLabel(entry.from_status!)} → ${statusLabel(entry.to_status!)}`
                  : entry.kind === 'note'
                    ? 'Nota'
                    : entry.kind === 'group_added'
                      ? 'Grupo agregado'
                      : 'Grupo quitado'}
              </strong>
              {entry.body && <p>{entry.body}</p>}
            </li>
          ))}
      </ol>
    </section>
  )
}

function DeploymentDialog({
  application,
  onClose,
  onCreated,
}: {
  application: string
  onClose: () => void
  onCreated: () => void
}) {
  const [component, setComponent] = useState(
    application === 'notizap' ? 'backend' : 'frontend',
  )
  const [reference, setReference] = useState('')
  const [actor, setActor] = useState('')
  const [error, setError] = useState('')
  async function submit(event: React.FormEvent) {
    event.preventDefault()
    if (!reference.trim()) {
      setError('Ingresá una versión o commit.')
      return
    }
    try {
      await api.createDeployment({ application, component, version: reference, actor })
      onCreated()
    } catch {
      setError('No se pudo registrar el deploy.')
    }
  }
  return (
    <div className="dialog-backdrop" role="presentation">
      <form className="deploy-dialog" onSubmit={submit}>
        <span className="eyebrow">Marker manual</span>
        <h2>Registrar deploy</h2>
        <p>{applicationNames[application]} · producción</p>
        <label>
          Componente
          <select
            value={component}
            onChange={(event) => setComponent(event.target.value)}
          >
            <option value="frontend">Frontend</option>
            <option value="backend">Backend</option>
          </select>
        </label>
        <label>
          Versión o commit
          <input
            value={reference}
            onChange={(event) => setReference(event.target.value)}
            maxLength={160}
          />
        </label>
        <label>
          Responsable
          <input
            value={actor}
            onChange={(event) => setActor(event.target.value)}
            maxLength={160}
          />
        </label>
        {error && <p className="form-error">{error}</p>}
        <div>
          <button type="button" onClick={onClose}>
            Cancelar
          </button>
          <button className="primary-action" type="submit">
            Registrar
          </button>
        </div>
      </form>
    </div>
  )
}
