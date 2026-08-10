import type { ChangeEvent } from 'react'

export interface EventFilterValues {
  application: string
  severity: string
  state: string
  period: string
}

interface EventFiltersProps {
  values: EventFilterValues
  onChange: (name: keyof EventFilterValues, value: string) => void
}

export function EventFilters({ values, onChange }: EventFiltersProps) {
  function handleChange(event: ChangeEvent<HTMLSelectElement>) {
    onChange(event.target.name as keyof EventFilterValues, event.target.value)
  }

  return (
    <section className="filters" aria-label="Filtros de eventos">
      <select
        name="application"
        aria-label="Aplicación"
        value={values.application}
        onChange={handleChange}
      >
        <option value="">Todas las aplicaciones</option>
        <option value="farmami">Farmami</option>
        <option value="wheels_house">Wheels House</option>
        <option value="prensap">Prensap</option>
        <option value="notizap">Notizap</option>
      </select>

      <select
        name="severity"
        aria-label="Severidad"
        value={values.severity}
        onChange={handleChange}
      >
        <option value="">Toda severidad</option>
        <option value="critical">critical</option>
        <option value="high">high</option>
        <option value="medium">medium</option>
        <option value="low">low</option>
        <option value="pending">pending</option>
      </select>

      <select
        name="state"
        aria-label="Estado"
        value={values.state}
        onChange={handleChange}
      >
        <option value="">Todo estado</option>
        <option value="actionable">Accionable</option>
        <option value="noise">Ruido</option>
        <option value="pending">Pendiente</option>
      </select>

      <select
        name="period"
        aria-label="Período"
        value={values.period}
        onChange={handleChange}
      >
        <option value="24h">24 horas</option>
        <option value="168h">7 días</option>
        <option value="720h">30 días</option>
        <option value="2160h">90 días</option>
      </select>
    </section>
  )
}
