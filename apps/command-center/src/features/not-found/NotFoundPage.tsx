import { Link } from 'react-router'

export function NotFoundPage() {
  return (
    <main className="not-found">
      <span className="eyebrow">ERROR / 404</span>
      <h1>Horizonte no encontrado.</h1>
      <p>La ruta solicitada no existe en Atalaya.</p>
      <Link to="/">Volver al faro</Link>
    </main>
  )
}
