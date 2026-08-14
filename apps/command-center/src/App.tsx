import { lazy, Suspense, useEffect, useState } from 'react'
import { Navigate, Outlet, Route, Routes, useNavigate, useParams } from 'react-router'
import { api, type Overview } from './api'
import { resolveApplication } from './catalog/applications'
import { ErrorState, LoadingState } from './components/feedback/FeedbackState'
import { CommandPaletteProvider } from './components/command/CommandPalette'
import { AppLayout } from './components/layout/AppLayout'
import { FineLayout } from './components/layout/FineLayout'
import { ArchitecturePage } from './features/architecture/ArchitecturePage'
import { LoginPage } from './features/auth/LoginPage'
import { EventDetailPage } from './features/events/EventDetailPage'
import { EventsPage } from './features/events/EventsPage'
import { LighthousePage } from './features/lighthouse/LighthousePage'
import { ApplicationPage } from './features/apps/ApplicationPage'
import { NotFoundPage } from './features/not-found/NotFoundPage'
import { OverviewPage } from './features/overview/OverviewPage'
import { ReportsPage } from './features/reports/ReportsPage'
import { StatusPage } from './features/status/StatusPage'
import { SystemPage } from './features/system/SystemPage'
import { clearViewMode, getInitialViewMode } from './lib/viewMode'
import './App.css'

const OperationsPage = lazy(() =>
  import('./features/operations/OperationsPage').then((module) => ({
    default: module.OperationsPage,
  })),
)

let sessionCheck: Promise<boolean> | null = null
function checkSession() {
  sessionCheck ??= api.session().then(
    () => true,
    () => false,
  )
  return sessionCheck
}

export default function App() {
  return (
    <Routes>
      <Route path="/status" element={<StatusPage />} />
      <Route element={<AuthGuard />}>
        <Route path="/" element={<HomeRoute />} />
        <Route path="/overview" element={<AppLayout />}>
          <Route index element={<OverviewRoute />} />
        </Route>
        <Route element={<ModeLayout />}>
          <Route path="/apps/:appSlug" element={<ApplicationRoute />} />
          <Route path="/events" element={<EventsRoute />} />
          <Route path="/events/:eventId" element={<EventDetailRoute />} />
          <Route
            path="/operations"
            element={
              <Suspense fallback={<LoadingState />}>
                <OperationsPage />
              </Suspense>
            }
          />
          <Route path="/reports" element={<ReportsPage />} />
          <Route path="/architecture/:appSlug?" element={<ArchitectureRoute />} />
          <Route path="/system" element={<SystemPage />} />
        </Route>
      </Route>
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  )
}

function AuthGuard() {
  const [authenticated, setAuthenticated] = useState<boolean | null>(null)
  const visualDemo =
    import.meta.env.DEV && new URLSearchParams(window.location.search).has('weather')

  useEffect(() => {
    void checkSession().then(setAuthenticated)
  }, [])

  if (visualDemo)
    return (
      <CommandPaletteProvider>
        <Outlet context={{ logout: async () => undefined }} />
      </CommandPaletteProvider>
    )
  if (authenticated === null) return <LoadingState />
  if (!authenticated) return <LoginPage onSuccess={() => setAuthenticated(true)} />

  async function logout() {
    clearViewMode()
    await api.logout().catch(() => undefined)
    sessionCheck = Promise.resolve(false)
    setAuthenticated(false)
  }

  return (
    <CommandPaletteProvider>
      <Outlet context={{ logout }} />
    </CommandPaletteProvider>
  )
}

function ModeLayout() {
  return getInitialViewMode() === 'classic' ? <AppLayout /> : <FineLayout />
}

function HomeRoute() {
  return getInitialViewMode() === 'classic' ? (
    <Navigate to="/overview" replace />
  ) : (
    <LighthousePage />
  )
}

function OverviewRoute() {
  const navigate = useNavigate()
  const [overview, setOverview] = useState<Overview | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    api
      .overview()
      .then(setOverview)
      .catch(() => setError('No se pudo cargar el estado general.'))
  }, [])

  if (error)
    return (
      <main className="page">
        <ErrorState message={error} />
      </main>
    )
  if (!overview)
    return (
      <main className="page">
        <LoadingState />
      </main>
    )
  return (
    <OverviewPage
      overview={overview}
      onSelectEvent={(id) => navigate(`/events/${id}`)}
      onSelectArchitecture={(slug) => navigate(`/architecture/${slug}`)}
    />
  )
}

function EventsRoute() {
  const navigate = useNavigate()
  return <EventsPage onSelectEvent={(id) => navigate(`/events/${id}`)} />
}

function EventDetailRoute() {
  const navigate = useNavigate()
  const { eventId } = useParams()
  return <EventDetailPage eventId={eventId!} onBack={() => navigate('/events')} />
}

function ArchitectureRoute() {
  const { appSlug } = useParams()
  if (!appSlug) return <Navigate to="/architecture/farmami" replace />
  const application = resolveApplication(appSlug)
  if (!application) return <NotFoundPage />
  if (application.slug !== appSlug.toLowerCase())
    return <Navigate to={`/architecture/${application.slug}`} replace />
  return <ArchitecturePage app={application} />
}

function ApplicationRoute() {
  const { appSlug } = useParams()
  const application = resolveApplication(appSlug)
  if (!application) return <NotFoundPage />
  if (application.slug !== appSlug?.toLowerCase())
    return <Navigate to={`/apps/${application.slug}`} replace />
  return <ApplicationPage app={application} />
}
