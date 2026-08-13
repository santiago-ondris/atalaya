import { Component, lazy, Suspense, type ErrorInfo, type ReactNode } from 'react'
import { useNavigate, useOutletContext, useSearchParams } from 'react-router'
import type { AuthOutletContext } from '../../components/layout/AppLayout'
import { setViewMode } from '../../lib/viewMode'

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

  function useClassicView() {
    setViewMode('classic')
    navigate('/overview')
  }

  return (
    <main className="lighthouse-page">
      <SceneBoundary fallback={<SceneFailure onClassic={useClassicView} />}>
        <Suspense fallback={<LighthouseCover message="Preparando el horizonte…" />}>
          <LighthouseScene still={still} />
        </Suspense>
      </SceneBoundary>
      <div className="lighthouse-overlay">
        <div className="lighthouse-brand">
          <span className="eyebrow">ATALAYA / FARO</span>
          <strong>Producción, a la vista.</strong>
        </div>
        <div className="lighthouse-actions">
          <button onClick={useClassicView}>Vista clásica</button>
          <button onClick={() => void logout()}>Cerrar sesión</button>
        </div>
      </div>
    </main>
  )
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
