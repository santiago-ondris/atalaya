interface SignalFlagProps {
  status?: string
}

export function SignalFlag({ status = 'healthy' }: SignalFlagProps) {
  return <span className={`signal signal-${status}`} aria-hidden="true" />
}
