import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Admin from '../pages/Admin'
import api from '../api/client'

vi.mock('../api/client', () => ({
  default: {
    get: vi.fn(),
    patch: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

const mockStats = { total_users: 42, total_blogs: 150, pending_reports: 3 }
const mockUsers = [
  { id: 'u1', username: 'alice', role: 'user', verified: true },
  { id: 'u2', username: 'bob', role: 'moderator', verified: true },
]
const mockReports: never[] = []

function wrap() {
  vi.mocked(api.get).mockImplementation((url: string) => {
    if (url.includes('stats')) return Promise.resolve({ data: mockStats }) as never
    if (url.includes('users')) return Promise.resolve({ data: { users: mockUsers } }) as never
    if (url.includes('reports')) return Promise.resolve({ data: { reports: mockReports } }) as never
    return Promise.resolve({ data: {} }) as never
  })
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <Admin />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('Admin page', () => {
  beforeEach(() => vi.clearAllMocks())

  it('renders stats tab by default', async () => {
    wrap()
    await waitFor(() => expect(screen.getByTestId('stat-total-users')).toBeInTheDocument())
    expect(screen.getByTestId('stat-total-users')).toHaveTextContent('42')
  })

  it('renders total blogs stat', async () => {
    wrap()
    await waitFor(() => expect(screen.getByTestId('stat-total-posts')).toHaveTextContent('150'))
  })

  it('renders pending reports stat', async () => {
    wrap()
    await waitFor(() => expect(screen.getByTestId('stat-pending-reports')).toHaveTextContent('3'))
  })

  it('switches to users tab and shows user rows', async () => {
    wrap()
    await userEvent.click(screen.getByTestId('admin-tab-users'))
    await waitFor(() => expect(screen.getAllByTestId('user-row')).toHaveLength(2))
  })

  it('shows promote button for regular users', async () => {
    wrap()
    await userEvent.click(screen.getByTestId('admin-tab-users'))
    await waitFor(() => expect(screen.getByTestId('promote-btn')).toBeInTheDocument())
  })

  it('switches to reports tab and shows empty state', async () => {
    wrap()
    await userEvent.click(screen.getByTestId('admin-tab-reports'))
    await waitFor(() => expect(screen.getByText(/no pending reports/i)).toBeInTheDocument())
  })

  it('calls api.patch when promote button is clicked', async () => {
    vi.mocked(api.patch).mockResolvedValue({ data: {} } as never)
    wrap()
    await userEvent.click(screen.getByTestId('admin-tab-users'))
    await waitFor(() => screen.getByTestId('promote-btn'))
    await userEvent.click(screen.getByTestId('promote-btn'))
    await waitFor(() => expect(api.patch).toHaveBeenCalledWith('/admin/users/u1', { role: 'moderator' }))
  })

  it('renders report rows when reports exist', async () => {
    vi.mocked(api.get).mockImplementation((url: string) => {
      if (url.includes('stats')) return Promise.resolve({ data: mockStats }) as never
      if (url.includes('users')) return Promise.resolve({ data: { users: mockUsers } }) as never
      if (url.includes('reports')) return Promise.resolve({
        data: {
          reports: [{
            id: 'r1', reason: 'spam content', reporter: { username: 'alice' },
            blog_id: 'b1', created_at: '2026-05-30T12:00:00Z',
          }],
        },
      }) as never
      return Promise.resolve({ data: {} }) as never
    })
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={qc}>
        <MemoryRouter>
          <Admin />
        </MemoryRouter>
      </QueryClientProvider>
    )
    await userEvent.click(screen.getByTestId('admin-tab-reports'))
    await waitFor(() => expect(screen.getByTestId('report-row')).toBeInTheDocument())
    expect(screen.getByText(/"spam content"/)).toBeInTheDocument()
  })
})
