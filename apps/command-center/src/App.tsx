import { lazy, Suspense, useEffect, useState } from 'react'
import { api, type Overview } from './api'
import { AppLayout, type View } from './components/layout/AppLayout'
import { ErrorState, LoadingState } from './components/feedback/FeedbackState'
import { ArchitecturePage } from './features/architecture/ArchitecturePage'
import { LoginPage } from './features/auth/LoginPage'
import { EventDetailPage } from './features/events/EventDetailPage'
import { EventsPage } from './features/events/EventsPage'
import { OverviewPage } from './features/overview/OverviewPage'
import { ReportsPage } from './features/reports/ReportsPage'
import { SystemPage } from './features/system/SystemPage'
import './App.css'

const OperationsPage = lazy(() =>
  import('./features/operations/OperationsPage').then((module) => ({
    default: module.OperationsPage,
  })),
)

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null)
  const [activeView, setActiveView] = useState<View>('overview')
  const [overview, setOverview] = useState<Overview | null>(null)
  const [selectedEventId, setSelectedEventId] = useState<string | null>(null)
  const [selectedArchApp, setSelectedArchApp] = useState<string | null>(null)
  const [overviewError, setOverviewError] = useState('')

  useEffect(() => {
    api
      .session()
      .then(() => setIsAuthenticated(true))
      .catch(() => setIsAuthenticated(false))
  }, [])

  useEffect(() => {
    if (!isAuthenticated) return

    api
      .overview()
      .then(setOverview)
      .catch((error: Error) => {
        if (error.message === 'unauthorized') {
          setIsAuthenticated(false)
          return
        }

        setOverviewError('No se pudo cargar el estado general.')
      })
  }, [isAuthenticated])

  if (isAuthenticated === null) {
    return <LoadingState />
  }

  if (!isAuthenticated) {
    return <LoginPage onSuccess={() => setIsAuthenticated(true)} />
  }

  async function handleLogout() {
    await api.logout()
    setIsAuthenticated(false)
  }

  function handleNavigation(view: View) {
    setActiveView(view)
    setSelectedEventId(null)
  }

  function handleOpenArchitecture(appSlug: string) {
    setSelectedArchApp(appSlug)
    setActiveView('architecture')
    setSelectedEventId(null)
  }

  function closeEventDetail() {
    setSelectedEventId(null)
    setActiveView('events')
  }

  return (
    <AppLayout view={activeView} onNavigate={handleNavigation} onLogout={handleLogout}>
      {renderContent()}
    </AppLayout>
  )

  function renderContent() {
    if (selectedEventId) {
      return <EventDetailPage eventId={selectedEventId} onBack={closeEventDetail} />
    }

    if (activeView === 'events') {
      return <EventsPage onSelectEvent={setSelectedEventId} />
    }

    if (activeView === 'architecture') {
      return <ArchitecturePage initialAppSlug={selectedArchApp} />
    }

    if (activeView === 'system') {
      return <SystemPage />
    }

    if (activeView === 'reports') {
      return <ReportsPage />
    }

    if (activeView === 'operations') {
      return (
        <Suspense fallback={<LoadingState />}>
          <OperationsPage />
        </Suspense>
      )
    }

    if (overviewError) {
      return (
        <main className="page">
          <ErrorState message={overviewError} />
        </main>
      )
    }

    if (!overview) {
      return (
        <main className="page">
          <LoadingState />
        </main>
      )
    }

    return (
      <OverviewPage
        onSelectArchitecture={handleOpenArchitecture}
        onSelectEvent={setSelectedEventId}
        overview={overview}
      />
    )
  }
}

export default App
