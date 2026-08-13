export { applicationNames } from '../catalog/applications'

const statusLabels: Record<string, string> = {
  ok: 'Operativo',
  healthy: 'Operativo',
  degraded: 'Degradado',
  error: 'Error',
  unconfigured: 'Sin configurar',
  actionable: 'Accionable',
  noise: 'Ruido',
  pending: 'Pendiente',
}

export function formatDateTime(value?: string): string {
  if (!value) return 'Sin registro'

  return new Intl.DateTimeFormat('es-AR', {
    dateStyle: 'short',
    timeStyle: 'medium',
  }).format(new Date(value))
}

export function formatStatus(value: string): string {
  return statusLabels[value] ?? value
}
