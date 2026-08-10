import { useState } from 'react'
import type { FormEvent } from 'react'
import { api } from '../../api'
import { SignalFlag } from '../../components/status/SignalFlag'

interface LoginPageProps {
  onSuccess: () => void
}

export function LoginPage({ onSuccess }: LoginPageProps) {
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setIsSubmitting(true)
    setError('')

    try {
      await api.login(password)
      onSuccess()
    } catch {
      setError('La contraseña no coincide con la guardia autorizada.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <main className="login">
      <div className="chart-field" />
      <div className="bearing" />

      <section className="login-copy">
        <span className="eyebrow">ATLY / ACCESO PRIVADO</span>
        <h1>La guardia empieza acá.</h1>
        <p>Ingresá al puesto de observación de producción.</p>

        <form onSubmit={handleSubmit}>
          <label htmlFor="password">Contraseña de guardia</label>
          <div className="login-row">
            <input
              id="password"
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoFocus
              autoComplete="current-password"
            />
            <button disabled={isSubmitting}>
              {isSubmitting ? 'Verificando…' : 'Entrar al puesto'}
            </button>
          </div>
          {error && (
            <div className="form-error" role="alert">
              {error}
            </div>
          )}
        </form>

        <div className="watch-note">
          <SignalFlag />
          Sesión privada · cookie segura HttpOnly
        </div>
      </section>
    </main>
  )
}
