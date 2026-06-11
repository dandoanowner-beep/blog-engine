import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import Forum from '../pages/Forum'

describe('Forum placeholder page (FR-CR002-004)', () => {
  it('renders the coming soon message', () => {
    render(
      <MemoryRouter>
        <Forum />
      </MemoryRouter>
    )
    expect(screen.getByText(/coming soon/i)).toBeInTheDocument()
  })
})
