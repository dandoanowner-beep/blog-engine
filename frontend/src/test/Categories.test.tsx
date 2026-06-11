import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Categories from '../pages/Categories'
import { blogsApi } from '../api/blogs'
import type { BlogCard } from '../types'

vi.mock('../api/blogs', () => ({
  blogsApi: {
    getCategories: vi.fn(),
    getArticlesFeed: vi.fn(),
  },
}))

const makeBlog = (id: string, title: string): BlogCard => ({
  id,
  title,
  excerpt: 'An excerpt',
  author: { id: 'u1', username: 'chubeunu' },
  read_time_min: 3,
  tags: [],
  like_count: 0,
  dislike_count: 0,
  comment_count: 0,
  privacy: 'public',
  published_at: '2026-06-10T00:00:00Z',
})

function wrap(ui: React.ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>{ui}</MemoryRouter>
    </QueryClientProvider>
  )
}

describe('Categories page (FR-CR002-003)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(blogsApi.getCategories).mockResolvedValue({
      data: {
        categories: [
          { id: 'c1', name: 'Tutorials', slug: 'tutorials', blog_count: 4 },
          { id: 'c2', name: 'Travel', slug: 'travel', blog_count: 0 },
        ],
      },
    } as never)
    vi.mocked(blogsApi.getArticlesFeed).mockResolvedValue({
      data: { blogs: [makeBlog('b1', 'Filtered Post')], total: 1, page: 1, per_page: 9 },
    } as never)
  })

  it('renders categories with article counts', async () => {
    wrap(<Categories />)
    await waitFor(() => expect(screen.getByText('Tutorials')).toBeInTheDocument())
    expect(screen.getByText('Travel')).toBeInTheDocument()
    expect(screen.getByText('4')).toBeInTheDocument()
  })

  it('clicking a category loads its filtered articles', async () => {
    wrap(<Categories />)
    await userEvent.click(await screen.findByTestId('category-tutorials'))
    await waitFor(() => expect(blogsApi.getArticlesFeed).toHaveBeenCalledWith(1, 'tutorials'))
    await waitFor(() => expect(screen.getByText('Filtered Post')).toBeInTheDocument())
  })

  it('shows empty state when no categories exist', async () => {
    vi.mocked(blogsApi.getCategories).mockResolvedValue({ data: { categories: [] } } as never)
    wrap(<Categories />)
    await waitFor(() => expect(screen.getByText(/no categories yet/i)).toBeInTheDocument())
  })
})
