import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import VerifyEmail from '../pages/VerifyEmail'
import { authApi } from '../api/auth'

vi.mock('../api/auth', () => ({
  authApi: { verifyEmail: vi.fn() },
}))

function wrap(query = '') {
  return render(
    <MemoryRouter initialEntries={[`/auth/verify${query}`]}>
      <VerifyEmail />
    </MemoryRouter>
  )
}

describe('VerifyEmail page', () => {
  beforeEach(() => vi.clearAllMocks())

  it('shows loading state initially', () => {
    vi.mocked(authApi.verifyEmail).mockReturnValue(new Promise(() => {}))
    wrap('?token=abc123')
    expect(screen.getByText(/verifying/i)).toBeInTheDocument()
  })

  it('shows success state after valid token', async () => {
    vi.mocked(authApi.verifyEmail).mockResolvedValue({} as never)
    wrap('?token=abc123')
    await waitFor(() => expect(screen.getByText(/email verified/i)).toBeInTheDocument())
  })

  it('shows error state when token is missing', async () => {
    wrap('')
    await waitFor(() => expect(screen.getByText(/verification failed/i)).toBeInTheDocument())
  })

  it('shows error state when API call fails', async () => {
    vi.mocked(authApi.verifyEmail).mockRejectedValue(new Error('expired'))
    wrap('?token=bad')
    await waitFor(() => expect(screen.getByText(/verification failed/i)).toBeInTheDocument())
  })

  it('shows sign in link on success', async () => {
    vi.mocked(authApi.verifyEmail).mockResolvedValue({} as never)
    wrap('?token=abc123')
    await waitFor(() => expect(screen.getByRole('link', { name: /sign in/i })).toBeInTheDocument())
  })
})
