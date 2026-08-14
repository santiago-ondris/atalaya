import {
  Component,
  lazy,
  Suspense,
  useState,
  type ErrorInfo,
  type ReactNode,
} from 'react'
import { Link, useNavigate, useOutletContext, useSearchParams } from 'react-router'
import type { AuthOutletContext } from '../../components/layout/AppLayout'
import { CommandPaletteTrigger } from '../../components/command/CommandPalette'
import { setViewMode } from '../../lib/viewMode'
import { findDestination, lighthouseDestinations } from './destinations'
import type { VisualSeverity } from './lighthouseState'
import { demoWeather, useLighthouseHealth } from './useLighthouseHealth'

const LighthouseScene = lazy(() =>
  import('./scene/LighthouseScene').then((module) => ({
    default: module.LighthouseScene,
  })),
)

export function LighthousePage() {
  const { logout } = useOutletContext<AuthOutletContext>()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const still = searchParams.get('scene') === 'still'
  const health = useLighthouseHealth()
  const weatherOverride = demoWeather(searchParams.get('weather'))
  const visualState = weatherOverride
    ? {
        ...health,
        weather: weatherOverride,
        applications: Object.fromEntries(
          Object.entries(health.applications).map(([slug, application]) => [
            slug,
            {
              ...application,
              severity: weatherSeverity(weatherOverride),
              freshness: 'fresh' as const,
            },
          ]),
        ) as typeof health.applications,
      }
    : health
  const [activeDestination, setActiveDestination] = useState<string | null>(null)

  function useClassicView() {
    setViewMode('classic')
    navigate('/overview')
  }

  return (
    <main className="lighthouse-page">
      <SceneBoundary fallback={<SceneFailure onClassic={useClassicView} />}>
        <Suspense fallback={<LighthouseCover message="Preparando el horizonte…" />}>
          <LighthouseScene
            still={still}
            visualState={visualState}
            activeDestination={activeDestination}
            onDestinationChange={setActiveDestination}
            onNavigate={(route) => navigate(route)}
          />
        </Suspense>
      </SceneBoundary>
      <div className="lighthouse-overlay">
        <div className="lighthouse-brand">
          <span className="eyebrow">ATALAYA / FARO</span>
          <strong>Producción, a la vista.</strong>
        </div>
        <div className="lighthouse-actions">
          <CommandPaletteTrigger />
          <button onClick={useClassicView}>Vista clásica</button>
          <button onClick={() => void logout()}>Cerrar sesión</button>
        </div>
      </div>
      <nav className="lighthouse-destinations" aria-label="Destinos del faro">
        {lighthouseDestinations.map((destination) => {
          const application =
            destination.kind === 'window'
              ? visualState.applications[
                  destination.id as keyof typeof visualState.applications
                ]
              : null
          const system = destination.id === 'system' ? visualState.system : null
          const status = application ?? system
          const accessibleLabel = status
            ? `${destination.label}: ${severityLabel(status.severity)}${
                status.freshness === 'stale' ? ', datos desactualizados' : ''
              }`
            : destination.label
          return (
            <Link
              key={destination.id}
              to={destination.route}
              aria-label={accessibleLabel}
              onFocus={() => setActiveDestination(destination.id)}
              onBlur={() => setActiveDestination(null)}
            >
              {accessibleLabel}
            </Link>
          )
        })}
      </nav>
      {activeDestination && (
        <div className="destination-label" aria-hidden="true">
          {destinationStatusLabel(activeDestination, visualState)}
        </div>
      )}
    </main>
  )
}

function severityLabel(severity: VisualSeverity) {
  return severity === 'green'
    ? 'operativo'
    : severity === 'red'
      ? 'interrupción mayor'
      : 'estado degradado o desconocido'
}

function weatherSeverity(weather: 'clear' | 'mist' | 'storm'): VisualSeverity {
  return weather === 'clear' ? 'green' : weather === 'mist' ? 'yellow' : 'red'
}

function destinationStatusLabel(
  id: string,
  visualState: ReturnType<typeof useLighthouseHealth>,
) {
  const destination = findDestination(id)
  if (!destination) return ''
  const status =
    destination.kind === 'window'
      ? visualState.applications[id as keyof typeof visualState.applications]
      : id === 'system'
        ? visualState.system
        : null
  if (!status) return destination.label
  return `${destination.label} · ${severityLabel(status.severity)}${
    status.freshness === 'stale' ? ' · datos desactualizados' : ''
  }`
}

export function LighthouseCover({ message }: { message: string }) {
  return (
    <section className="lighthouse-cover" aria-live="polite">
      <span className="eyebrow">ATALAYA / FARO</span>
      <div className="lighthouse-mark" aria-hidden="true">
        <span />
      </div>
      <h1>Producción, a la vista.</h1>
      <p>{message}</p>
    </section>
  )
}

function SceneFailure({ onClassic }: { onClassic: () => void }) {
  return (
    <section className="lighthouse-cover scene-failure" role="alert">
      <span className="eyebrow">ATALAYA / RECUPERACIÓN</span>
      <h1>El faro no pudo encenderse.</h1>
      <p>La operación sigue disponible en la vista clásica.</p>
      <button className="primary-action" onClick={onClassic}>
        Abrir vista clásica
      </button>
    </section>
  )
}

interface SceneBoundaryProps {
  children: ReactNode
  fallback: ReactNode
}
interface SceneBoundaryState {
  failed: boolean
}

export class SceneBoundary extends Component<SceneBoundaryProps, SceneBoundaryState> {
  state = { failed: false }

  static getDerivedStateFromError(): SceneBoundaryState {
    return { failed: true }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('Lighthouse scene failed', error, info)
  }

  render() {
    return this.state.failed ? this.props.fallback : this.props.children
  }
}
