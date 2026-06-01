import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Editor from '../pages/Editor'
import { blogsApi } from '../api/blogs'

vi.mock('../api/blogs', () => ({
  blogsApi: {
    createBlog: vi.fn(),
    updateBlog: vi.fn(),
    getBlog: vi.fn(),
    uploadImage: vi.fn(),
  },
}))

vi.mock('@tiptap/react', () => ({
  useEditor: () => ({
    chain: () => ({ focus: () => ({ toggleBold: () => ({ run: vi.fn() }), toggleItalic: () => ({ run: vi.fn() }), toggleCode: () => ({ run: vi.fn() }), toggleCodeBlock: () => ({ run: vi.fn() }), setImage: () => ({ run: vi.fn() }), setContent: vi.fn() }) }),
    getHTML: () => '<p>Content</p>',
    commands: { setContent: vi.fn() },
  }),
  EditorContent: ({ 'data-testid': dtid }: { 'data-testid'?: string }) => (
    <div data-testid={dtid ?? 'editor-content'}>Editor</div>
  ),
}))
vi.mock('@tiptap/starter-kit', () => ({ default: {} }))
vi.mock('@tiptap/extension-image', () => ({ default: {} }))
vi.mock('@tiptap/extension-link', () => ({ default: { configure: () => ({}) } }))
vi.mock('@tiptap/extension-code-block-lowlight', () => ({ default: { configure: () => ({}) } }))
vi.mock('lowlight', () => ({ common: {}, createLowlight: () => ({}) }))

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async (importOriginal) => {
  const mod = await importOriginal<typeof import('react-router-dom')>()
  return { ...mod, useNavigate: () => mockNavigate }
})

function wrap(path = '/editor') {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[path]}>
        <Routes>
          <Route path="/editor" element={<Editor />} />
          <Route path="/editor/:id" element={<Editor />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('Editor page', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(blogsApi.createBlog).mockResolvedValue({ data: { id: 'b1', title: 'T' } } as never)
  })

  it('renders title input', () => {
    wrap()
    expect(screen.getByTestId('title-input')).toBeInTheDocument()
  })

  it('renders editor area', () => {
    wrap()
    expect(screen.getByTestId('editor-content')).toBeInTheDocument()
  })

  it('renders privacy selector with public as default', () => {
    wrap()
    expect(screen.getByTestId('privacy-select')).toHaveValue('public')
  })

  it('shows error when publishing without title', async () => {
    wrap()
    await userEvent.click(screen.getByTestId('publish-btn'))
    expect(screen.getByRole('alert')).toHaveTextContent(/title is required/i)
  })

  it('calls createBlog with title and content on publish', async () => {
    wrap()
    await userEvent.type(screen.getByTestId('title-input'), 'My First Post')
    await userEvent.click(screen.getByTestId('publish-btn'))
    await waitFor(() => expect(blogsApi.createBlog).toHaveBeenCalledWith(
      expect.objectContaining({ title: 'My First Post', status: 'published' })
    ))
  })

  it('navigates to blog detail after publish', async () => {
    wrap()
    await userEvent.type(screen.getByTestId('title-input'), 'New Post')
    await userEvent.click(screen.getByTestId('publish-btn'))
    await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith('/blog/b1'))
  })

  it('renders 3 privacy options', () => {
    wrap()
    const select = screen.getByTestId('privacy-select')
    expect(select.querySelectorAll('option')).toHaveLength(3)
  })

  it('saves as draft when save draft button is clicked', async () => {
    wrap()
    await userEvent.type(screen.getByTestId('title-input'), 'Draft Post')
    await userEvent.click(screen.getByRole('button', { name: /save draft/i }))
    await waitFor(() => expect(blogsApi.createBlog).toHaveBeenCalledWith(
      expect.objectContaining({ status: 'draft' })
    ))
  })

  it('loads existing blog data in edit mode', async () => {
    vi.mocked(blogsApi.getBlog).mockResolvedValue({
      data: {
        id: 'b1', title: 'Old Title', content: '<p>old</p>',
        privacy: 'friend_only', tags: [{ id: 't1', name: 'go', slug: 'go' }],
        thumbnail_url: 'http://cdn/thumb.jpg',
        author: { id: 'u1', username: 'alice' }, read_time_min: 2,
        like_count: 0, dislike_count: 0, comment_count: 0, published_at: '2026-05-30T00:00:00Z',
      },
    } as never)
    wrap('/editor/b1')
    await waitFor(() => expect(screen.getByTestId('title-input')).toHaveValue('Old Title'))
    expect(screen.getByTestId('privacy-select')).toHaveValue('friend_only')
  })

  it('calls updateBlog instead of createBlog in edit mode', async () => {
    vi.mocked(blogsApi.getBlog).mockResolvedValue({
      data: {
        id: 'b1', title: 'Old Title', content: '<p>old</p>',
        privacy: 'public', tags: [], thumbnail_url: undefined,
        author: { id: 'u1', username: 'alice' }, read_time_min: 1,
        like_count: 0, dislike_count: 0, comment_count: 0, published_at: '2026-05-30T00:00:00Z',
      },
    } as never)
    vi.mocked(blogsApi.updateBlog).mockResolvedValue({ data: { id: 'b1' } } as never)
    wrap('/editor/b1')
    await waitFor(() => expect(screen.getByTestId('title-input')).toHaveValue('Old Title'))
    await userEvent.click(screen.getByTestId('publish-btn'))
    await waitFor(() => expect(blogsApi.updateBlog).toHaveBeenCalledWith('b1', expect.objectContaining({ status: 'published' })))
    expect(blogsApi.createBlog).not.toHaveBeenCalled()
  })

  it('shows error message when save fails', async () => {
    vi.mocked(blogsApi.createBlog).mockRejectedValue(new Error('Network error'))
    wrap()
    await userEvent.type(screen.getByTestId('title-input'), 'My Post')
    await userEvent.click(screen.getByTestId('publish-btn'))
    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent(/failed to save/i))
  })

  it('uploads image file when file input changes', async () => {
    vi.mocked(blogsApi.uploadImage).mockResolvedValue({ data: { url: 'http://cdn/photo.jpg' } } as never)
    wrap()
    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
    const file = new File(['img'], 'photo.png', { type: 'image/png' })
    await userEvent.upload(fileInput, file)
    await waitFor(() => expect(blogsApi.uploadImage).toHaveBeenCalledWith(file))
  })

  it('shows image upload error when upload fails', async () => {
    vi.mocked(blogsApi.uploadImage).mockRejectedValue(new Error('upload failed'))
    wrap()
    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
    const file = new File(['img'], 'photo.png', { type: 'image/png' })
    await userEvent.upload(fileInput, file)
    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent(/image upload failed/i))
  })

  it('toolbar buttons invoke editor chain commands', async () => {
    wrap()
    await userEvent.click(screen.getByRole('button', { name: 'B' }))
    await userEvent.click(screen.getByRole('button', { name: 'I' }))
    await userEvent.click(screen.getByRole('button', { name: '`' }))
    await userEvent.click(screen.getByRole('button', { name: '```' }))
  })
})
