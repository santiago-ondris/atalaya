export type Severity = 'critical' | 'high' | 'medium' | 'low' | 'pending'
export type EventState = 'actionable' | 'noise' | 'pending'

export interface EventSummary {
  id: string
  source_event_id: string
  application: string
  component: string
  environment: string
  error_type: string
  message: string
  occurred_at: string
  ingested_at: string
  severity: Severity
  state: EventState
  occurrence_count: number
}
export interface IntegrationStatus {
  id: string
  application: string
  component: string
  display_name: string
  source: string
  project: string
  enabled: boolean
  environments: string[]
  status: string
  last_attempt_at?: string
  last_success_at?: string
  last_error?: string
}
export interface AppStatus {
  slug: string
  display_name: string
  status: string
  last_success_at?: string
}
export interface Overview {
  applications: AppStatus[]
  integrations: IntegrationStatus[]
  recent_events: EventSummary[]
  event_total: number
  generated_at: string
}
export interface Interpretation {
  summary: string
  explanation: string
  severity: Severity
  actionable: boolean
  suggested_actions: string[]
  model: string
  total_tokens: number
  estimated_cost_usd?: number
  latency_ms: number
}
export interface EventOccurrence {
  id: string
  source_event_id: string
  occurred_at: string
  message: string
}
export interface EventDetail extends EventSummary {
  fingerprint: string
  stack_trace?: string
  release?: string
  metadata: Record<string, unknown>
  interpretation?: Interpretation
  occurrences: EventOccurrence[]
}
export interface DailyReportApplication {
  application: string
  display_name: string
  activity_count?: number
  activity_kind: 'sessions' | 'page_views'
  activity_source: string
  activity_status: 'available' | 'unavailable'
  activity_error?: string
  error_count: number
  occurrence_count: number
  severity_counts: Record<'critical' | 'high' | 'medium' | 'low', number>
  actionable_count: number
}
export interface DailyReport {
  id: string
  date: string
  timezone: string
  period_start: string
  period_end: string
  status: 'collecting' | 'pending' | 'processing' | 'sent' | 'expired'
  attempts: number
  last_error?: string
  created_at: string
  sent_at?: string
  expired_at?: string
  applications: DailyReportApplication[]
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/json', ...init?.headers },
  })
  if (response.status === 401) throw new Error('unauthorized')
  if (!response.ok)
    throw new Error(
      (await response.json().catch(() => null))?.error ??
        `request failed (${response.status})`,
    )
  return response.status === 204 ? (undefined as T) : response.json()
}

export const api = {
  session: () => request<{ authenticated: boolean }>('/api/v1/session'),
  login: (password: string) =>
    request('/api/v1/session', { method: 'POST', body: JSON.stringify({ password }) }),
  logout: () => request<void>('/api/v1/session', { method: 'DELETE' }),
  overview: () => request<Overview>('/api/v1/overview'),
  integrations: () =>
    request<{ integrations: IntegrationStatus[] }>('/api/v1/integrations'),
  events: (query: URLSearchParams) =>
    request<{ events: EventSummary[]; total: number; limit: number; offset: number }>(
      `/api/v1/events?${query}`,
    ),
  event: (id: string) => request<EventDetail>(`/api/v1/events/${id}`),
  reports: () => request<{ reports: DailyReport[] }>('/api/v1/reports?limit=30'),
}
