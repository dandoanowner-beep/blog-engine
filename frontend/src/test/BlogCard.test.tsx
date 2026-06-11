import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import BlogCard from '../components/BlogCard'
import type { BlogCard as BlogCardType } from '../types'

vi.mock('react-i18next', () => ({
  useTranslation: vi.fn(() => ({
    t: (key: string) => ({
      'blog.minRead': 'min read',
      'blog.signInToReadMore': 'Sign in to read more →',
    } as Record<string, string>)[key] ?? key,
    i18n: { changeLanguage: vi.fn(), language: 'vi' },
  })),
  initReactI18next: { type: '3rdParty', init: vi.fn() },
}))

import { useTranslation } from 'react-i18next'

const blog: BlogCardType = {
  id: 'b1',
  title: 'Hello World',
  excerpt: 'A short excerpt about something interesting.',
  author: { id: 'u1', username: 'alice' },
  read_time_min: 5,
  tags: [{ id: 't1', name: 'Go', slug: 'go' }],
  like_count: 10,
  dislike_count: 2,
  comment_count: 3,
  privacy: 'public',
  published_at: '2026-05-30T00:00:00Z',
}

function wrap(ui: React.ReactElement) {
  return render(<MemoryRouter>{ui}</MemoryRouter>)
}

describe('BlogCard', () => {
  it('renders title', () => {
    wrap(<BlogCard blog={blog} />)
    expect(screen.getByText('Hello World')).toBeInTheDocument()
  })

  it('renders author username', () => {
    wrap(<BlogCard blog={blog} />)
    expect(screen.getByText('alice')).toBeInTheDocument()
  })

  it('renders tag', () => {
    wrap(<BlogCard blog={blog} />)
    expect(screen.getByText('Go')).toBeInTheDocument()
  })

  it('renders read time', () => {
    wrap(<BlogCard blog={blog} />)
    expect(screen.getByText('5 min read')).toBeInTheDocument()
  })

  it('renders like count', () => {
    wrap(<BlogCard blog={blog} />)
    expect(screen.getByText('👍 10')).toBeInTheDocument()
  })

  it('does NOT render guest prompt when not partial', () => {
    wrap(<BlogCard blog={blog} />)
    expect(screen.queryByText(/sign in to read more/i)).not.toBeInTheDocument()
  })

  it('renders guest prompt when blog is partial', () => {
    wrap(<BlogCard blog={{ ...blog, partial: true }} />)
    expect(screen.getByText(/sign in to read more/i)).toBeInTheDocument()
  })

  it('renders thumbnail when provided', () => {
    wrap(<BlogCard blog={{ ...blog, thumbnail_url: 'https://img.test/t.jpg' }} />)
    expect(screen.getByRole('img')).toHaveAttribute('src', 'https://img.test/t.jpg')
  })

  it('has testid blog-card', () => {
    wrap(<BlogCard blog={blog} />)
    expect(screen.getByTestId('blog-card')).toBeInTheDocument()
  })
})

// ════════════════════════════════════════════════════════════
// BlogCard — i18n language switching
// ════════════════════════════════════════════════════════════

describe('BlogCard — language switching', () => {
  const i18nBlog: BlogCardType = {
    ...blog,
    title_en: 'Hello World in English',
    translation_status: 'done',
  }

  beforeEach(() => {
    vi.mocked(useTranslation).mockReturnValue({
      t: (key: string) => ({ 'blog.minRead': 'min read', 'blog.signInToReadMore': 'Sign in to read more →' } as Record<string, string>)[key] ?? key,
      i18n: { changeLanguage: vi.fn(), language: 'vi' },
    } as never)
  })

  afterEach(() => vi.clearAllMocks())

  it('shows VI title by default when language is vi', () => {
    wrap(<BlogCard blog={i18nBlog} />)
    expect(screen.getByText('Hello World')).toBeInTheDocument()
  })

  it('shows EN title when language is en and translation_status is done', () => {
    vi.mocked(useTranslation).mockReturnValue({
      t: (key: string) => key,
      i18n: { changeLanguage: vi.fn(), language: 'en' },
    } as never)
    wrap(<BlogCard blog={i18nBlog} />)
    expect(screen.getByText('Hello World in English')).toBeInTheDocument()
  })

  it('falls back to VI title when language is en but translation_status is not done', () => {
    vi.mocked(useTranslation).mockReturnValue({
      t: (key: string) => key,
      i18n: { changeLanguage: vi.fn(), language: 'en' },
    } as never)
    wrap(<BlogCard blog={{ ...blog, translation_status: 'none' }} />)
    expect(screen.getByText('Hello World')).toBeInTheDocument()
  })
})
