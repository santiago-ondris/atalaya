import './App.css'

function App() {
  return (
    <main>
      <header>
        <p className="eyebrow">ATLY / SYSTEM LOG</p>
        <h1>Atalaya</h1>
        <p className="lede">
          El command center está listo para recibir su primera señal.
        </p>
      </header>

      <section aria-labelledby="foundation-title">
        <div className="section-heading">
          <span>00</span>
          <h2 id="foundation-title">Fundación del sistema</h2>
        </div>

        <div className="service-grid">
          <article>
            <span className="signal signal-ok" aria-label="Configurado" />
            <p className="service-kind">Go</p>
            <h3>Watchman</h3>
            <p>API, reglas operativas y persistencia.</p>
          </article>
          <article>
            <span className="signal signal-ok" aria-label="Configurado" />
            <p className="service-kind">Python</p>
            <h3>Interpreter</h3>
            <p>Interpretación estructurada mediante LLM.</p>
          </article>
          <article>
            <span className="signal signal-attention" aria-label="En construcción" />
            <p className="service-kind">PostgreSQL</p>
            <h3>Bitácora</h3>
            <p>Eventos, jobs e historial operacional.</p>
          </article>
        </div>
      </section>
    </main>
  )
}

export default App
