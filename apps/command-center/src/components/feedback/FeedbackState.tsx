import { SignalFlag } from '../status/SignalFlag'

interface FeedbackStateProps {
  message: string
}

export function LoadingState() {
  return (
    <div className="state">
      <span className="spinner" />
      Consultando bitácora…
    </div>
  )
}

export function EmptyState({ message }: FeedbackStateProps) {
  return (
    <div className="state">
      <SignalFlag status="unconfigured" />
      {message}
    </div>
  )
}

export function ErrorState({ message }: FeedbackStateProps) {
  return (
    <div className="state state-error" role="alert">
      <SignalFlag status="error" />
      {message}
    </div>
  )
}
