export const applications = [
  {
    slug: 'farmami',
    aliases: [],
    displayName: 'Farmami',
    badge: 'Sentry',
    stack: 'Vue / Laravel / MySQL',
    deploy: 'Railway + Vercel',
    htmlUrl: '/diagrams/arquitectura-farmami.html',
    brandColor: '#D85A30',
  },
  {
    slug: 'wheels_house',
    aliases: ['wheelshouse'],
    displayName: 'Wheels House',
    badge: 'Sentry',
    stack: 'React / Node.js / PostgreSQL',
    deploy: 'Railway + Vercel',
    htmlUrl: '/diagrams/arquitectura-wheelshouse.html',
    brandColor: '#BF247A',
  },
  {
    slug: 'prensap',
    aliases: ['prensapp'],
    displayName: 'Prensap',
    badge: 'Sentry',
    stack: 'Next.js / PostgreSQL / Cloudflare',
    deploy: 'Cloudflare Pages + Railway',
    htmlUrl: '/diagrams/arquitectura-prensap.html',
    brandColor: '#D7FF3F',
  },
  {
    slug: 'notizap',
    aliases: [],
    displayName: 'Notizap',
    badge: 'App Insights (KQL)',
    stack: 'Node.js / Express / Azure',
    deploy: 'Azure Web App + GitHub Actions',
    htmlUrl: '/diagrams/arquitectura-notizap.html',
    brandColor: '#B695BF',
  },
  {
    slug: 'atalaya',
    aliases: ['watchman', 'interpreter', 'command_center'],
    displayName: 'Atalaya',
    badge: 'Meta-observabilidad',
    stack: 'Go / Python / React / PostgreSQL',
    deploy: 'Railway + Cloudflare Pages',
    htmlUrl: '/diagrams/arquitectura-atalaya.html',
    brandColor: '#4750A8',
  },
] as const

export type Application = (typeof applications)[number]
export type ApplicationSlug = Application['slug']

export function resolveApplication(value: string | undefined): Application | undefined {
  if (!value) return undefined
  const normalized = value.toLowerCase()
  return applications.find(
    (application) =>
      application.slug === normalized ||
      application.aliases.some((alias) => alias === normalized),
  )
}

export const applicationNames: Record<string, string> = Object.fromEntries(
  applications.flatMap((application) => [
    [application.slug, application.displayName],
    ...application.aliases.map((alias) => [alias, application.displayName]),
  ]),
)
