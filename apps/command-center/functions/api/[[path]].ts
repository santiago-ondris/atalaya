interface Env {
  WATCHMAN_ORIGIN: string
}
export const onRequest: PagesFunction<Env> = async ({ request, env }) => {
  if (!env.WATCHMAN_ORIGIN)
    return new Response('WATCHMAN_ORIGIN is not configured', { status: 503 })
  const incoming = new URL(request.url)
  let target: URL
  try {
    target = new URL(incoming.pathname + incoming.search, env.WATCHMAN_ORIGIN.trim())
  } catch {
    return new Response('WATCHMAN_ORIGIN must be a valid absolute URL', { status: 503 })
  }

  const hasBody = request.method !== 'GET' && request.method !== 'HEAD'
  return fetch(target.toString(), {
    method: request.method,
    headers: request.headers,
    body: hasBody ? request.body : undefined,
    redirect: 'manual',
  })
}
