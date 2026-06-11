import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import Author from '../pages/Author'

// OQ-005 (2026-06-11): the Author page is STATIC — owner's design instructions
// pending. No editing UI may exist on this page for anyone.
describe('Author page (static — design pending)', () => {
  it('renders the page heading', () => {
    render(
      <MemoryRouter>
        <Author />
      </MemoryRouter>
    )
    expect(screen.getByRole('heading', { name: /author/i })).toBeInTheDocument()
  })

  it('renders the interim placeholder', () => {
    render(
      <MemoryRouter>
        <Author />
      </MemoryRouter>
    )
    expect(screen.getByText(/coming soon/i)).toBeInTheDocument()
  })

  it('has NO edit button — for anyone', () => {
    render(
      <MemoryRouter>
        <Author />
      </MemoryRouter>
    )
    expect(screen.queryByTestId('edit-about-btn')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /edit/i })).not.toBeInTheDocument()
  })
})
