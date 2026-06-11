import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Layout from '../components/Layout'
import { useAuthStore } from '../store/auth'

vi.mock('../store/auth', () => ({ useAuthStore: vi.fn() }))
vi.mock('../api/social', () => ({
  socialApi: { getNotifications: vi.fn().mockResolvedValue({ data: { notifications: [], unread_count: 0, total: 0 } }) },
}))

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async (importOriginal) => {
  const mod = await importOriginal<typeof import('react-router-dom')>()
  return { ...mod, useNavigate: () => mockNavigate }
})

function wrap(user: object | null) {
  vi.mocked(useAuthStore).mockReturnValue({ user, logout: vi.fn().mockResolvedValue(undefined) } as never)
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <Layout><div data-testid="child">Content</div></Layout>
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('Layout component', () => {
  beforeEach(() => vi.clearAllMocks())

  it('renders children', () => {
    wrap(null)
    expect(screen.getByTestId('child')).toBeInTheDocument()
  })

  it('shows sign in and get started links for guests', () => {
    wrap(null)
    expect(screen.getByRole('link', { name: /sign in/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /get started/i })).toBeInTheDocument()
  })

  // CR-001 personal-blog pivot: only the owner writes
  it('shows Write button for the owner', () => {
    wrap({ id: '1', username: 'chubeunu', role: 'owner', avatar_url: null })
    expect(screen.getByRole('link', { name: /write/i })).toBeInTheDocument()
  })

  it('does NOT show Write button for regular users', () => {
    wrap({ id: '1', username: 'alice', role: 'user', avatar_url: null })
    expect(screen.queryByRole('link', { name: /write/i })).not.toBeInTheDocument()
  })

  it('shows the Articles nav link instead of Explore', () => {
    wrap(null)
    expect(screen.getByRole('link', { name: /articles/i })).toBeInTheDocument()
    expect(screen.queryByRole('link', { name: /explore/i })).not.toBeInTheDocument()
  })

  // CR-002: four new header sections, visible to everyone
  it('shows Portfolio, Author, Categories and Forums nav links', () => {
    wrap(null)
    expect(screen.getByRole('link', { name: /portfolio/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /author/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /categories/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /forums/i })).toBeInTheDocument()
  })

  it('shows Admin link for admin users', () => {
    wrap({ id: '1', username: 'admin', role: 'admin', avatar_url: null })
    expect(screen.getByRole('link', { name: /admin/i })).toBeInTheDocument()
  })

  it('does NOT show Admin link for regular users', () => {
    wrap({ id: '1', username: 'alice', role: 'user', avatar_url: null })
    expect(screen.queryByRole('link', { name: /admin/i })).not.toBeInTheDocument()
  })

  it('calls logout and navigates on Logout click', async () => {
    const mockLogout = vi.fn().mockResolvedValue(undefined)
    vi.mocked(useAuthStore).mockReturnValue({ user: { id: '1', username: 'alice', role: 'user' }, logout: mockLogout } as never)
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={qc}>
        <MemoryRouter>
          <Layout><div /></Layout>
        </MemoryRouter>
      </QueryClientProvider>
    )
    await userEvent.click(screen.getByRole('button', { name: /logout/i }))
    await waitFor(() => expect(mockLogout).toHaveBeenCalled())
  })
})
