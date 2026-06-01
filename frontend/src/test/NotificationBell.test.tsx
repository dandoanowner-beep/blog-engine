import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import NotificationBell from '../components/NotificationBell'
import { socialApi } from '../api/social'

vi.mock('../api/social', () => ({
  socialApi: {
    getNotifications: vi.fn(),
    markAllRead: vi.fn(),
  },
}))

const emptyNotifications = { notifications: [], unread_count: 0, total: 0 }

function wrap() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <NotificationBell />
    </QueryClientProvider>
  )
}

describe('NotificationBell', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(socialApi.getNotifications).mockResolvedValue({ data: emptyNotifications } as never)
    vi.mocked(socialApi.markAllRead).mockResolvedValue({} as never)
  })

  it('renders bell button', () => {
    wrap()
    expect(screen.getByTestId('notification-bell')).toBeInTheDocument()
  })

  it('does not show badge when unread count is 0', async () => {
    wrap()
    await waitFor(() => expect(socialApi.getNotifications).toHaveBeenCalled())
    expect(screen.queryByText('0')).not.toBeInTheDocument()
  })

  it('shows unread count badge when there are unread notifications', async () => {
    vi.mocked(socialApi.getNotifications).mockResolvedValue({
      data: { notifications: [], unread_count: 3, total: 3 },
    } as never)
    wrap()
    await waitFor(() => expect(screen.getByText('3')).toBeInTheDocument())
  })

  it('shows 9+ when unread count exceeds 9', async () => {
    vi.mocked(socialApi.getNotifications).mockResolvedValue({
      data: { notifications: [], unread_count: 15, total: 15 },
    } as never)
    wrap()
    await waitFor(() => expect(screen.getByText('9+')).toBeInTheDocument())
  })

  it('opens dropdown on bell click', async () => {
    wrap()
    await userEvent.click(screen.getByTestId('notification-bell'))
    expect(screen.getByTestId('notification-list')).toBeInTheDocument()
  })

  it('closes dropdown on second bell click', async () => {
    wrap()
    await userEvent.click(screen.getByTestId('notification-bell'))
    expect(screen.getByTestId('notification-list')).toBeInTheDocument()
    await userEvent.click(screen.getByTestId('notification-bell'))
    expect(screen.queryByTestId('notification-list')).not.toBeInTheDocument()
  })

  it('shows no notifications message when list is empty', async () => {
    wrap()
    await userEvent.click(screen.getByTestId('notification-bell'))
    await waitFor(() => expect(screen.getByText(/no notifications/i)).toBeInTheDocument())
  })

  it('renders notification items in list', async () => {
    vi.mocked(socialApi.getNotifications).mockResolvedValue({
      data: {
        notifications: [
          { id: 'n1', type: 'new_like', actor: { id: 'u2', username: 'bob' }, read: false, created_at: '2026-05-31T00:00:00Z' },
        ],
        unread_count: 1,
        total: 1,
      },
    } as never)
    wrap()
    await userEvent.click(screen.getByTestId('notification-bell'))
    await waitFor(() => expect(screen.getByText('bob')).toBeInTheDocument())
  })

  it('calls markAllRead when mark all read button is clicked', async () => {
    wrap()
    await userEvent.click(screen.getByTestId('notification-bell'))
    await userEvent.click(screen.getByRole('button', { name: /mark all read/i }))
    expect(socialApi.markAllRead).toHaveBeenCalled()
  })
})
