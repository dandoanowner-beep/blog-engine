import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Feed from '../pages/Feed'
import { blogsApi } from '../api/blogs'
import type { BlogCard } from '../types'

vi.mock('../api/blogs', () => ({
  blogsApi: {
    getArticlesFeed: vi.fn(),
  },
}))

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

// CR-001 personal-blog pivot: one article feed, no Explore/Following tabs.
describe('Feed page (single article feed)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(blogsApi.getArticlesFeed).mockResolvedValue({
      data: { blogs: [makeBlog('b1', 'First Post')], total: 1, page: 1, per_page: 9 },
    } as never)
  })

  it('renders article cards', async () => {
    wrap(<Feed />)
    await waitFor(() => expect(screen.getByText('First Post')).toBeInTheDocument())
  })

  it('does not render Explore/Following tabs', () => {
    wrap(<Feed />)
    expect(screen.queryByTestId('tab-explore')).not.toBeInTheDocument()
    expect(screen.queryByTestId('tab-following')).not.toBeInTheDocument()
  })

  it('fetches page 1 of the article feed on mount', async () => {
    wrap(<Feed />)
    await waitFor(() => expect(blogsApi.getArticlesFeed).toHaveBeenCalledWith(1))
  })

  it('shows empty state when there are no posts', async () => {
    vi.mocked(blogsApi.getArticlesFeed).mockResolvedValue({
      data: { blogs: [], total: 0, page: 1, per_page: 9 },
    } as never)
    wrap(<Feed />)
    await waitFor(() => expect(screen.getByText(/no posts yet/i)).toBeInTheDocument())
  })
})
