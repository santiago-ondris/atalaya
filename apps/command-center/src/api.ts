export type Severity = 'critical' | 'high' | 'medium' | 'low' | 'pending'
export type EventState = 'actionable' | 'noise' | 'pending'

export interface EventSummary {
  id: string
  error_group_id: string
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

export type IncidentStatus = 'investigating' | 'resolved' | 'noise'
export interface ErrorGroupSummary {
  id: string
  application: string
  component: string
  error_type: string
  message: string
  occurrence_count: number
  last_seen_at: string
}
export interface IncidentEntry {
  id: string
  kind: 'note' | 'status_change' | 'group_added' | 'group_removed'
  body?: string
  from_status?: IncidentStatus
  to_status?: IncidentStatus
  created_at: string
}
export interface Incident {
  id: string
  application: string
  title: string
  status: IncidentStatus
  created_at: string
  updated_at: string
  closed_at?: string
  groups?: ErrorGroupSummary[]
  entries?: IncidentEntry[]
}
export interface Deployment {
  id: string
  application: string
  component: string
  environment: string
  provider: 'railway' | 'github_actions' | 'manual'
  external_id: string
  version?: string
  commit_sha?: string
  commit_url?: string
  actor?: string
  source_url?: string
  deployed_at: string
  created_at: string
}
export interface TimelineBucket {
  start: string
  error_count: number
}
export interface OperationsTimeline {
  application: string
  from: string
  to: string
  bucket_seconds: number
  timeline: {
    buckets: TimelineBucket[]
    deployments: Deployment[]
    incidents: Incident[]
  }
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
export type AvailabilityStatus = 'unknown' | 'operational' | 'degraded' | 'major_outage'
export interface PublicStatus {
  status: AvailabilityStatus
  generated_at: string
  applications: Array<{
    slug: string
    display_name: string
    status: AvailabilityStatus
    components: Array<{
      name: string
      status: 'unknown' | 'up' | 'down'
      last_checked_at?: string
    }>
    uptime_30_days: number | null
    last_checked_at?: string
  }>
  incidents: Array<{
    id: string
    application: string
    title: string
    message: string
    status: IncidentStatus
    published_at: string
    resolved_at?: string
  }>
}
export interface SystemHealth {
  status: string
  generated_at: string
  signals: Array<{ name: string; status: string; detail?: string; last_seen_at?: string }>
  queues: Record<string, number>
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
  publicStatus: () => request<PublicStatus>('/api/v1/public/status'),
  session: () => request<{ authenticated: boolean }>('/api/v1/session'),
  login: (password: string) =>
    request('/api/v1/session', { method: 'POST', body: JSON.stringify({ password }) }),
  logout: () => request<void>('/api/v1/session', { method: 'DELETE' }),
  overview: () => request<Overview>('/api/v1/overview'),
  integrations: () =>
    request<{ integrations: IntegrationStatus[] }>('/api/v1/integrations'),
  systemHealth: () => request<SystemHealth>('/api/v1/system/health'),
  events: (query: URLSearchParams) =>
    request<{ events: EventSummary[]; total: number; limit: number; offset: number }>(
      `/api/v1/events?${query}`,
    ),
  event: (id: string) => request<EventDetail>(`/api/v1/events/${id}`),
  reports: () => request<{ reports: DailyReport[] }>('/api/v1/reports?limit=30'),
  timeline: (application: string, component: string, period: string) =>
    request<OperationsTimeline>(
      `/api/v1/operations/timeline?application=${encodeURIComponent(application)}&component=${encodeURIComponent(component)}&period=${period}`,
    ),
  incident: (id: string) => request<Incident>(`/api/v1/incidents/${id}`),
  createIncident: (title: string, errorGroupIds: string[]) =>
    request<Incident>('/api/v1/incidents', {
      method: 'POST',
      body: JSON.stringify({ title, error_group_ids: errorGroupIds }),
    }),
  addIncidentNote: (id: string, body: string) =>
    request(`/api/v1/incidents/${id}/notes`, {
      method: 'POST',
      body: JSON.stringify({ body }),
    }),
  changeIncidentStatus: (id: string, status: IncidentStatus, note: string) =>
    request(`/api/v1/incidents/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status, note }),
    }),
  errorGroups: (application: string, query = '') =>
    request<{ groups: ErrorGroupSummary[] }>(
      `/api/v1/error-groups?application=${encodeURIComponent(application)}&q=${encodeURIComponent(query)}&limit=25`,
    ),
  addIncidentGroup: (incidentId: string, groupId: string) =>
    request<void>(`/api/v1/incidents/${incidentId}/groups/${groupId}`, {
      method: 'POST',
    }),
  removeIncidentGroup: (incidentId: string, groupId: string) =>
    request<void>(`/api/v1/incidents/${incidentId}/groups/${groupId}`, {
      method: 'DELETE',
    }),
  createDeployment: (input: {
    application: string
    component: string
    version?: string
    commit_sha?: string
    deployed_at?: string
    actor?: string
  }) =>
    request<Deployment>('/api/v1/deployments', {
      method: 'POST',
      body: JSON.stringify(input),
    }),
  costs: () => request<CostSummary>('/api/v1/system/costs'),
}

export interface ApplicationCostBreakdown {
  application: string
  total_tokens: number
  estimated_cost_usd: number
  request_count: number
}

export interface ModelCostBreakdown {
  model: string
  total_tokens: number
  estimated_cost_usd: number
  request_count: number
}

export interface CostSummary {
  total_cost_usd: number
  monthly_cost_usd: number
  monthly_budget_usd: number
  budget_used_percent: number
  total_tokens: number
  input_tokens: number
  output_tokens: number
  total_requests: number
  average_latency_ms: number
  by_application: ApplicationCostBreakdown[]
  by_model: ModelCostBreakdown[]
}

