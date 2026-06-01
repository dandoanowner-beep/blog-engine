import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import ResetPassword from '../pages/ResetPassword'
import { authApi } from '../api/auth'

vi.mock('../api/auth', () => ({
  authApi: { forgotPassword: vi.fn(), resetPassword: vi.fn() },
}))

function wrap(query = '') {
  return render(
    <MemoryRouter initialEntries={[`/auth/forgot-password${query}`]}>
      <ResetPassword />
    </MemoryRouter>
  )
}

describe('ResetPassword page', () => {
  beforeEach(() => vi.clearAllMocks())

  it('renders email field in forgot-password mode', () => {
    wrap()
    expect(screen.getByLabelText('Email')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /send reset link/i })).toBeInTheDocument()
  })

  it('calls forgotPassword on submit', async () => {
    vi.mocked(authApi.forgotPassword).mockResolvedValue({} as never)
    wrap()
    await userEvent.type(screen.getByLabelText('Email'), 'alice@test.com')
    await userEvent.click(screen.getByRole('button', { name: /send reset link/i }))
    expect(authApi.forgotPassword).toHaveBeenCalledWith('alice@test.com')
  })

  it('shows confirmation after forgot-password submit', async () => {
    vi.mocked(authApi.forgotPassword).mockResolvedValue({} as never)
    wrap()
    await userEvent.type(screen.getByLabelText('Email'), 'alice@test.com')
    await userEvent.click(screen.getByRole('button', { name: /send reset link/i }))
    await waitFor(() => expect(screen.getByText(/check your email/i)).toBeInTheDocument())
  })

  it('shows error when forgotPassword fails', async () => {
    vi.mocked(authApi.forgotPassword).mockRejectedValue(new Error('fail'))
    wrap()
    await userEvent.type(screen.getByLabelText('Email'), 'alice@test.com')
    await userEvent.click(screen.getByRole('button', { name: /send reset link/i }))
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument())
  })

  it('renders password field in reset-password mode (with token)', () => {
    wrap('?token=abc123')
    expect(screen.getByLabelText('New password')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /update password/i })).toBeInTheDocument()
  })

  it('calls resetPassword with token and new password', async () => {
    vi.mocked(authApi.resetPassword).mockResolvedValue({} as never)
    wrap('?token=abc123')
    await userEvent.type(screen.getByLabelText('New password'), 'newpassword1')
    await userEvent.click(screen.getByRole('button', { name: /update password/i }))
    expect(authApi.resetPassword).toHaveBeenCalledWith('abc123', 'newpassword1')
  })

  it('shows success after password reset', async () => {
    vi.mocked(authApi.resetPassword).mockResolvedValue({} as never)
    wrap('?token=abc123')
    await userEvent.type(screen.getByLabelText('New password'), 'newpassword1')
    await userEvent.click(screen.getByRole('button', { name: /update password/i }))
    await waitFor(() => expect(screen.getByText(/password updated/i)).toBeInTheDocument())
  })
})
