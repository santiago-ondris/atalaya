import { applications } from '../../catalog/applications'

export type DestinationKind = 'window' | 'patrol' | 'cargo' | 'mail' | 'buoy'

export interface LighthouseDestination {
  id: string
  label: string
  route: string
  kind: DestinationKind
}

export const lighthouseDestinations: LighthouseDestination[] = [
  ...applications.map((application) => ({
    id: application.slug,
    label: application.displayName,
    route: `/apps/${application.slug}`,
    kind: 'window' as const,
  })),
  { id: 'events', label: 'Eventos', route: '/events', kind: 'patrol' },
  { id: 'operations', label: 'Bitácora', route: '/operations', kind: 'cargo' },
  { id: 'reports', label: 'Reportes', route: '/reports', kind: 'mail' },
  { id: 'system', label: 'Estado del sistema', route: '/system', kind: 'buoy' },
]

export function findDestination(id: string) {
  return lighthouseDestinations.find((destination) => destination.id === id)
}
