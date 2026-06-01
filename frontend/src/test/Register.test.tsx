import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import Register from '../pages/Register'
import { useAuthStore } from '../store/auth'

vi.mock('../store/auth', () => ({ useAuthStore: vi.fn() }))
vi.mock('../api/auth', () => ({ authApi: { googleLogin: vi.fn() } }))

const mockRegister = vi.fn()

describe('Register page', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useAuthStore).mockReturnValue({ register: mockRegister, loading: false } as never)
  })

  function renderRegister() {
    return render(<MemoryRouter><Register /></MemoryRouter>)
  }

  it('renders all three form fields', () => {
    renderRegister()
    expect(screen.getByLabelText('Email')).toBeInTheDocument()
    expect(screen.getByLabelText('Username')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
  })

  it('calls register with correct args', async () => {
    mockRegister.mockResolvedValue(undefined)
    renderRegister()

    await userEvent.type(screen.getByLabelText('Email'), 'bob@test.com')
    await userEvent.type(screen.getByLabelText('Username'), 'bob123')
    await userEvent.type(screen.getByLabelText('Password'), 'password!')
    await userEvent.click(screen.getByRole('button', { name: /create account/i }))

    expect(mockRegister).toHaveBeenCalledWith('bob@test.com', 'bob123', 'password!')
  })

  it('shows email check screen after successful registration', async () => {
    mockRegister.mockResolvedValue(undefined)
    renderRegister()

    await userEvent.type(screen.getByLabelText('Email'), 'bob@test.com')
    await userEvent.type(screen.getByLabelText('Username'), 'bob')
    await userEvent.type(screen.getByLabelText('Password'), 'password!')
    await userEvent.click(screen.getByRole('button', { name: /create account/i }))

    await waitFor(() => expect(screen.getByText(/check your email/i)).toBeInTheDocument())
  })

  it('shows error on registration failure', async () => {
    mockRegister.mockRejectedValue({ response: { data: { error: 'email already taken' } } })
    renderRegister()

    await userEvent.type(screen.getByLabelText('Email'), 'dup@test.com')
    await userEvent.type(screen.getByLabelText('Username'), 'dupuser')
    await userEvent.type(screen.getByLabelText('Password'), 'password1')
    await userEvent.click(screen.getByRole('button', { name: /create account/i }))

    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('email already taken'))
  })

  it('link to sign in exists', () => {
    renderRegister()
    expect(screen.getByRole('link', { name: /sign in/i })).toHaveAttribute('href', '/auth/login')
  })
})
