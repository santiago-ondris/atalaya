import { applications } from '../../catalog/applications'

export type DestinationKind = 'window' | 'patrol' | 'cargo' | 'mail' | 'buoy'

export interface LighthouseDestination {
  id: string
  label: string
  route: string
  kind: DestinationKind
  group: 'Aplicaciones' | 'Operación'
  context: string
  keywords: string[]
}

export const lighthouseDestinations: LighthouseDestination[] = [
  ...applications.map((application) => ({
    id: application.slug,
    label: application.displayName,
    route: `/apps/${application.slug}`,
    kind: 'window' as const,
    group: 'Aplicaciones' as const,
    context:
      application.kind === 'platform'
        ? 'Plataforma de observabilidad'
        : application.stack,
    keywords: [application.slug, ...application.aliases],
  })),
  {
    id: 'events',
    label: 'Eventos',
    route: '/events',
    kind: 'patrol',
    group: 'Operación',
    context: 'Alertas e incidentes',
    keywords: ['errores', 'incidentes'],
  },
  {
    id: 'operations',
    label: 'Bitácora',
    route: '/operations',
    kind: 'cargo',
    group: 'Operación',
    context: 'Operaciones y despliegues',
    keywords: ['operaciones', 'despliegues', 'colas'],
  },
  {
    id: 'reports',
    label: 'Reportes',
    route: '/reports',
    kind: 'mail',
    group: 'Operación',
    context: 'Informes operativos',
    keywords: ['informes'],
  },
  {
    id: 'system',
    label: 'Estado del sistema',
    route: '/system',
    kind: 'buoy',
    group: 'Operación',
    context: 'Salud de la plataforma',
    keywords: ['salud'],
  },
]

export function findDestination(id: string) {
  return lighthouseDestinations.find((destination) => destination.id === id)
}
