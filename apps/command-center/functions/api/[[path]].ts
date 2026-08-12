interface Env {
  WATCHMAN_ORIGIN: string
}
export const onRequest: PagesFunction<Env> = async ({ request, env }) => {
  if (!env.WATCHMAN_ORIGIN)
    return new Response('WATCHMAN_ORIGIN is not configured', { status: 503 })
  const incoming = new URL(request.url)
  const target = new URL(incoming.pathname + incoming.search, env.WATCHMAN_ORIGIN)
  return fetch(new Request(target.toString(), request))
}
