import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import BlogDetail from '../pages/BlogDetail'
import { blogsApi } from '../api/blogs'
import { useAuthStore } from '../store/auth'
import type { Blog } from '../types'

vi.mock('../api/blogs', () => ({
  blogsApi: {
    getBlog: vi.fn(),
    getComments: vi.fn(),
    react: vi.fn(),
    createComment: vi.fn(),
    deleteBlog: vi.fn(),
    share: vi.fn(),
    copyLink: vi.fn(),
  },
}))
vi.mock('../store/auth', () => ({ useAuthStore: vi.fn() }))

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async (importOriginal) => {
  const mod = await importOriginal<typeof import('react-router-dom')>()
  return { ...mod, useNavigate: () => mockNavigate }
})

const blog: Blog = {
  id: 'b1',
  title: 'Test Post',
  excerpt: 'Excerpt',
  content: '<p>Full content</p>',
  author: { id: 'u1', username: 'alice' },
  read_time_min: 4,
  tags: [],
  like_count: 5,
  dislike_count: 1,
  comment_count: 0,
  privacy: 'public',
  published_at: '2026-05-30T00:00:00Z',
}

function wrap(blogId = 'b1') {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[`/blog/${blogId}`]}>
        <Routes>
          <Route path="/blog/:id" element={<BlogDetail />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('BlogDetail page', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useAuthStore).mockReturnValue({ user: null } as never)
    vi.mocked(blogsApi.getBlog).mockResolvedValue({ data: blog } as never)
    vi.mocked(blogsApi.getComments).mockResolvedValue({ data: { items: [], total: 0 } } as never)
  })

  it('renders blog title', async () => {
    wrap()
    await waitFor(() => expect(screen.getByTestId('blog-title')).toHaveTextContent('Test Post'))
  })

  it('renders blog content HTML', async () => {
    wrap()
    await waitFor(() => expect(screen.getByTestId('blog-content')).toBeInTheDocument())
  })

  it('shows guest prompt for partial content', async () => {
    vi.mocked(blogsApi.getBlog).mockResolvedValue({ data: { ...blog, partial: true } } as never)
    wrap()
    await waitFor(() => expect(screen.getByTestId('guest-prompt')).toBeInTheDocument())
  })

  it('does NOT show guest prompt for full content', async () => {
    wrap()
    await waitFor(() => expect(screen.queryByTestId('guest-prompt')).not.toBeInTheDocument())
  })

  it('shows comment input for logged-in user', async () => {
    vi.mocked(useAuthStore).mockReturnValue({ user: { id: 'u2', username: 'bob', role: 'user' } } as never)
    wrap()
    await waitFor(() => expect(screen.getByTestId('comment-input')).toBeInTheDocument())
  })

  it('does NOT show delete button for non-author non-mod', async () => {
    vi.mocked(useAuthStore).mockReturnValue({ user: { id: 'u99', username: 'stranger', role: 'user' } } as never)
    wrap()
    await waitFor(() => expect(screen.queryByTestId('delete-btn')).not.toBeInTheDocument())
  })

  it('shows delete button for blog author', async () => {
    vi.mocked(useAuthStore).mockReturnValue({ user: { id: 'u1', username: 'alice', role: 'user' } } as never)
    wrap()
    await waitFor(() => expect(screen.getByTestId('delete-btn')).toBeInTheDocument())
  })

  it('shows not found for missing blog', async () => {
    vi.mocked(blogsApi.getBlog).mockRejectedValue({ response: { status: 404 } })
    wrap()
    await waitFor(() => expect(screen.getByText(/post not found/i)).toBeInTheDocument())
  })

  it('submits a comment and clears the input on success', async () => {
    vi.mocked(useAuthStore).mockReturnValue({ user: { id: 'u2', username: 'bob', role: 'user' } } as never)
    vi.mocked(blogsApi.createComment).mockResolvedValue({ data: { id: 'c1', content: 'hello world' } } as never)
    wrap()
    await waitFor(() => screen.getByTestId('comment-input'))
    await userEvent.type(screen.getByTestId('comment-input'), 'hello world')
    await userEvent.click(screen.getByRole('button', { name: /^post$/i }))
    await waitFor(() => expect(blogsApi.createComment).toHaveBeenCalledWith('b1', 'hello world'))
    await waitFor(() => expect(screen.getByTestId('comment-input')).toHaveValue(''))
  })

  it('deletes the post and navigates home when author confirms', async () => {
    vi.mocked(useAuthStore).mockReturnValue({ user: { id: 'u1', username: 'alice', role: 'user' } } as never)
    vi.mocked(blogsApi.deleteBlog).mockResolvedValue({} as never)
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    wrap()
    await waitFor(() => screen.getByTestId('delete-btn'))
    await userEvent.click(screen.getByTestId('delete-btn'))
    await waitFor(() => expect(blogsApi.deleteBlog).toHaveBeenCalled())
    await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith('/'))
  })

  it('calls react mutation when like button is clicked', async () => {
    vi.mocked(useAuthStore).mockReturnValue({ user: { id: 'u2', username: 'bob', role: 'user' } } as never)
    vi.mocked(blogsApi.react).mockResolvedValue({} as never)
    wrap()
    await waitFor(() => screen.getByTestId('blog-title'))
    const likeBtn = document.querySelector('[aria-label="Like"]') as HTMLElement
    await userEvent.click(likeBtn)
    await waitFor(() => expect(blogsApi.react).toHaveBeenCalledWith('b1', 'like'))
  })

  it('renders thumbnail image when blog has thumbnail_url', async () => {
    vi.mocked(blogsApi.getBlog).mockResolvedValue({
      data: { ...blog, thumbnail_url: 'http://cdn/thumb.jpg' },
    } as never)
    wrap()
    await waitFor(() => expect(screen.getByRole('img', { name: 'Test Post' })).toHaveAttribute('src', 'http://cdn/thumb.jpg'))
  })
})
