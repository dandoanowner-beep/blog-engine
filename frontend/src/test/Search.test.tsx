import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Search from '../pages/Search'
import { socialApi } from '../api/social'
import type { SearchResults } from '../types'

vi.mock('../api/social', () => ({
  socialApi: { search: vi.fn() },
}))

const emptyResults: SearchResults = {
  query: 'test',
  blogs: { items: [], total: 0, page: 1 },
  users: { items: [], total: 0 },
  tags: { items: [], total: 0 },
}

function wrap(initialEntry = '/search') {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[initialEntry]}>
        <Search />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('Search page', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(socialApi.search).mockResolvedValue({ data: emptyResults } as never)
  })

  it('renders search input', () => {
    wrap()
    expect(screen.getByTestId('search-input')).toBeInTheDocument()
  })

  it('does not call API for empty query', () => {
    wrap()
    expect(socialApi.search).not.toHaveBeenCalled()
  })

  it('calls API when query is >= 2 chars (from URL param)', async () => {
    wrap('/search?q=go')
    await waitFor(() => expect(socialApi.search).toHaveBeenCalledWith('go', 1))
  })

  it('shows no results message when empty', async () => {
    wrap('/search?q=xyz')
    await waitFor(() => expect(screen.getByText(/no results for/i)).toBeInTheDocument())
  })

  it('renders blog section when blogs found', async () => {
    vi.mocked(socialApi.search).mockResolvedValue({
      data: {
        ...emptyResults,
        blogs: {
          items: [{
            id: 'b1', title: 'Go Generics', excerpt: 'e', author: { id: 'u1', username: 'alice' },
            read_time_min: 2, tags: [], like_count: 0, dislike_count: 0, comment_count: 0,
            privacy: 'public' as const, published_at: '2026-05-30T00:00:00Z',
          }],
          total: 1, page: 1,
        },
      },
    } as never)
    wrap('/search?q=go')
    await waitFor(() => expect(screen.getByText('Posts (1)')).toBeInTheDocument())
    expect(screen.getByText('Go Generics')).toBeInTheDocument()
  })

  it('updates query on form submit', async () => {
    wrap()
    await userEvent.type(screen.getByTestId('search-input'), 'react')
    await userEvent.click(screen.getByRole('button', { name: /search/i }))
    await waitFor(() => expect(socialApi.search).toHaveBeenCalledWith('react', 1))
  })

  it('renders users section when users are found', async () => {
    vi.mocked(socialApi.search).mockResolvedValue({
      data: {
        ...emptyResults,
        users: {
          items: [{
            id: 'u1', username: 'alice', bio: 'Dev',
            follower_count: 5, following_count: 2, friend_count: 1,
            viewer_relation: 'guest' as const,
          }],
          total: 1,
        },
      },
    } as never)
    wrap('/search?q=al')
    await waitFor(() => expect(screen.getByText('People (1)')).toBeInTheDocument())
    expect(screen.getByText('alice')).toBeInTheDocument()
  })

  it('renders tags section when tags are found', async () => {
    vi.mocked(socialApi.search).mockResolvedValue({
      data: {
        ...emptyResults,
        tags: {
          items: [{ id: 't1', name: 'react', slug: 'react' }],
          total: 1,
        },
      },
    } as never)
    wrap('/search?q=re')
    await waitFor(() => expect(screen.getByText('Tags (1)')).toBeInTheDocument())
    expect(screen.getByText('#react')).toBeInTheDocument()
  })
})
