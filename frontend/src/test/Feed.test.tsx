import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Feed from '../pages/Feed'
import { blogsApi } from '../api/blogs'
import { useAuthStore } from '../store/auth'
import type { BlogCard } from '../types'

vi.mock('../api/blogs', () => ({
  blogsApi: {
    getExploreFeed: vi.fn(),
    getFollowingFeed: vi.fn(),
  },
}))

vi.mock('../store/auth', () => ({ useAuthStore: vi.fn() }))

const makeBlog = (id: string, title: string): BlogCard => ({
  id,
  title,
  excerpt: 'An excerpt',
  author: { id: 'u1', username: 'alice' },
  read_time_min: 3,
  tags: [],
  like_count: 0,
  dislike_count: 0,
  comment_count: 0,
  privacy: 'public',
  published_at: '2026-05-30T00:00:00Z',
})

function wrap(ui: React.ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>{ui}</MemoryRouter>
    </QueryClientProvider>
  )
}

describe('Feed page', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useAuthStore).mockReturnValue({ user: null } as never)
    vi.mocked(blogsApi.getExploreFeed).mockResolvedValue({
      data: { blogs: [makeBlog('b1', 'First Post')], total: 1, page: 1, per_page: 9 },
    } as never)
  })

  it('renders Explore tab by default', () => {
    wrap(<Feed />)
    expect(screen.getByTestId('tab-explore')).toBeInTheDocument()
  })

  it('renders Following tab', () => {
    wrap(<Feed />)
    expect(screen.getByTestId('tab-following')).toBeInTheDocument()
  })

  it('renders blog cards from explore feed', async () => {
    wrap(<Feed />)
    await waitFor(() => expect(screen.getByText('First Post')).toBeInTheDocument())
  })

  it('shows login prompt on Following tab for guest', async () => {
    wrap(<Feed />)
    await userEvent.click(screen.getByTestId('tab-following'))
    await waitFor(() => expect(screen.getByText(/sign in to see posts/i)).toBeInTheDocument())
  })

  it('fetches following feed when user is logged in and Following tab active', async () => {
    vi.mocked(useAuthStore).mockReturnValue({ user: { id: 'u1', username: 'alice', role: 'user' } } as never)
    vi.mocked(blogsApi.getFollowingFeed).mockResolvedValue({
      data: { blogs: [makeBlog('b2', 'Following Post')], total: 1, page: 1 },
    } as never)

    wrap(<Feed />)
    await userEvent.click(screen.getByTestId('tab-following'))
    await waitFor(() => expect(blogsApi.getFollowingFeed).toHaveBeenCalled())
  })

  it('resets to page 1 on tab change', async () => {
    wrap(<Feed />)
    await userEvent.click(screen.getByTestId('tab-following'))
    await userEvent.click(screen.getByTestId('tab-explore'))
    expect(blogsApi.getExploreFeed).toHaveBeenCalledWith(1)
  })
})
