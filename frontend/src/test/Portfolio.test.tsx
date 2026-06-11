import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Portfolio from '../pages/Portfolio'
import { projectsApi } from '../api/projects'
import { useAuthStore } from '../store/auth'

vi.mock('../api/projects', () => ({
  projectsApi: {
    list: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
  },
}))

vi.mock('../store/auth', () => ({ useAuthStore: vi.fn() }))

const project = {
  id: 'p1',
  title: 'Blog Engine',
  description: '<p>A personal blog platform</p>',
  tech_stack: 'Go, React, PostgreSQL',
  repo_url: 'https://github.com/x/blog-engine',
  demo_url: '',
  thumbnail_url: '',
  sort_order: 0,
}

function wrap(ui: React.ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>{ui}</MemoryRouter>
    </QueryClientProvider>
  )
}

describe('Portfolio page (FR-CR002-001)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useAuthStore).mockReturnValue({ user: null } as never)
    vi.mocked(projectsApi.list).mockResolvedValue({ data: { projects: [project] } } as never)
  })

  it('renders project cards', async () => {
    wrap(<Portfolio />)
    await waitFor(() => expect(screen.getByText('Blog Engine')).toBeInTheDocument())
    expect(screen.getByText('Go, React, PostgreSQL')).toBeInTheDocument()
  })

  it('shows repo link when present', async () => {
    wrap(<Portfolio />)
    await waitFor(() => expect(screen.getByRole('link', { name: /code/i })).toHaveAttribute('href', project.repo_url))
  })

  it('does NOT show Add project for guests', async () => {
    wrap(<Portfolio />)
    await waitFor(() => expect(screen.getByText('Blog Engine')).toBeInTheDocument())
    expect(screen.queryByTestId('add-project-btn')).not.toBeInTheDocument()
  })

  it('does NOT show Add project for regular users', async () => {
    vi.mocked(useAuthStore).mockReturnValue({ user: { id: 'u1', username: 'alice', role: 'user' } } as never)
    wrap(<Portfolio />)
    await waitFor(() => expect(screen.getByText('Blog Engine')).toBeInTheDocument())
    expect(screen.queryByTestId('add-project-btn')).not.toBeInTheDocument()
  })

  it('owner can open the add form and create a project', async () => {
    vi.mocked(useAuthStore).mockReturnValue({ user: { id: 'o1', username: 'chubeunu', role: 'owner' } } as never)
    vi.mocked(projectsApi.create).mockResolvedValue({ data: { ...project, id: 'p2', title: 'New Project' } } as never)

    wrap(<Portfolio />)
    await userEvent.click(await screen.findByTestId('add-project-btn'))
    await userEvent.type(screen.getByTestId('project-title-input'), 'New Project')
    await userEvent.click(screen.getByTestId('project-save-btn'))

    await waitFor(() => expect(projectsApi.create).toHaveBeenCalled())
    expect(vi.mocked(projectsApi.create).mock.calls[0][0].title).toBe('New Project')
  })

  it('shows empty state when there are no projects', async () => {
    vi.mocked(projectsApi.list).mockResolvedValue({ data: { projects: [] } } as never)
    wrap(<Portfolio />)
    await waitFor(() => expect(screen.getByText(/no projects yet/i)).toBeInTheDocument())
  })
})
