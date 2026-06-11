import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import Login from '../pages/Login'
import { useAuthStore } from '../store/auth'

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async (importOriginal) => {
  const mod = await importOriginal<typeof import('react-router-dom')>()
  return { ...mod, useNavigate: () => mockNavigate }
})

vi.mock('../store/auth', () => ({
  useAuthStore: vi.fn(),
}))

vi.mock('../api/auth', () => ({
  authApi: { googleLogin: vi.fn() },
}))

const mockLogin = vi.fn()

describe('Login page', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useAuthStore).mockReturnValue({ login: mockLogin, loading: false } as never)
  })

  function renderLogin() {
    return render(<MemoryRouter><Login /></MemoryRouter>)
  }

  it('renders email and password inputs', () => {
    renderLogin()
    expect(screen.getByLabelText('Email')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
  })

  it('renders sign in button', () => {
    renderLogin()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('calls login with email and password on submit', async () => {
    mockLogin.mockResolvedValue(undefined)
    renderLogin()

    await userEvent.type(screen.getByLabelText('Email'), 'alice@test.com')
    await userEvent.type(screen.getByLabelText('Password'), 'secret123')
    await userEvent.click(screen.getByRole('button', { name: /sign in/i }))

    expect(mockLogin).toHaveBeenCalledWith('alice@test.com', 'secret123')
  })

  it('navigates to / on successful login', async () => {
    mockLogin.mockResolvedValue(undefined)
    renderLogin()

    await userEvent.type(screen.getByLabelText('Email'), 'a@b.com')
    await userEvent.type(screen.getByLabelText('Password'), 'pass')
    await userEvent.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith('/'))
  })

  it('shows error message on failed login', async () => {
    mockLogin.mockRejectedValue({ response: { status: 401, data: { error: 'Invalid email or password' } } })
    renderLogin()

    await userEvent.type(screen.getByLabelText('Email'), 'a@b.com')
    await userEvent.type(screen.getByLabelText('Password'), 'wrong')
    await userEvent.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent(/invalid email or password/i))
  })

  it('disables submit button while loading', () => {
    vi.mocked(useAuthStore).mockReturnValue({ login: mockLogin, loading: true } as never)
    renderLogin()
    expect(screen.getByRole('button', { name: /signing in/i })).toBeDisabled()
  })

  it('has link to register page', () => {
    renderLogin()
    expect(screen.getByRole('link', { name: /get started/i })).toHaveAttribute('href', '/auth/register')
  })
})
