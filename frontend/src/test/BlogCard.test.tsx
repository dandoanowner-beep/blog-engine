import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import BlogCard from '../components/BlogCard'
import type { BlogCard as BlogCardType } from '../types'

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
