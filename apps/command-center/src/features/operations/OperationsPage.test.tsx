import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import type { Deployment } from '../../api'
import { DeploymentMarker } from './OperationsPage'

afterEach(cleanup)

describe('DeploymentMarker', () => {
  it('shows deployment metadata and navigation links', () => {
    const deployment: Deployment = {
      id: 'deployment-1',
      application: 'prensap',
      component: 'frontend',
      environment: 'production',
      provider: 'github_actions',
      external_id: 'deployment-url',
      commit_sha: 'abcdef1234567890',
      commit_url: 'https://github.com/example/prensap/commit/abcdef1234567890',
      actor: 'santiago',
      source_url: 'https://deployment.pages.dev',
      deployed_at: '2026-08-11T12:00:00Z',
      created_at: '2026-08-11T12:00:01Z',
    }

    render(<DeploymentMarker deployment={deployment} />)

    expect(screen.getByText('production')).toBeInTheDocument()
    expect(screen.getByText('GitHub Actions')).toBeInTheDocument()
    expect(screen.getByText('frontend')).toBeInTheDocument()
    expect(screen.getByText('santiago')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Commit abcdef12' })).toHaveAttribute(
      'href',
      deployment.commit_url,
    )
    expect(screen.getByRole('link', { name: 'Abrir deployment' })).toHaveAttribute(
      'href',
      deployment.source_url,
    )
  })

  it('uses the version and omits unavailable optional links', () => {
    render(
      <DeploymentMarker
        deployment={{
          id: 'deployment-2',
          application: 'prensap',
          component: 'backend',
          environment: 'production',
          provider: 'railway',
          external_id: 'railway-1',
          version: 'railway-deployment-railway-1',
          deployed_at: '2026-08-11T12:00:00Z',
          created_at: '2026-08-11T12:00:01Z',
        }}
      />,
    )

    expect(screen.getByText('railway-deployment-railway-1')).toBeInTheDocument()
    expect(screen.getByText('Railway')).toBeInTheDocument()
    expect(screen.queryByRole('link')).not.toBeInTheDocument()
  })
})
