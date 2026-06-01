import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Profile from '../pages/Profile'
import { socialApi } from '../api/social'
import { useAuthStore } from '../store/auth'
import type { UserProfile } from '../types'

vi.mock('../api/social', () => ({
  socialApi: {
    getProfile: vi.fn(),
    follow: vi.fn(),
    unfollow: vi.fn(),
    updateProfile: vi.fn(),
  },
}))
vi.mock('../store/auth', () => ({ useAuthStore: vi.fn() }))

const aliceProfile: UserProfile = {
  id: 'u1',
  username: 'alice',
  bio: 'Developer',
  follower_count: 10,
  following_count: 5,
  friend_count: 3,
  viewer_relation: 'owner',
}

const strangerProfile: UserProfile = {
  ...aliceProfile,
  id: 'u2',
  username: 'bob',
  viewer_relation: 'stranger',
}

function wrap(username = 'alice') {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[`/profile/${username}`]}>
        <Routes>
          <Route path="/profile/:username" element={<Profile />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('Profile page', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useAuthStore).mockReturnValue({ user: { id: 'u1', username: 'alice', role: 'user' } } as never)
    vi.mocked(socialApi.getProfile).mockResolvedValue({ data: { user: aliceProfile } } as never)
  })

  it('renders username', async () => {
    wrap()
    await waitFor(() => expect(screen.getByTestId('profile-username')).toHaveTextContent('alice'))
  })

  it('renders follower count', async () => {
    wrap()
    await waitFor(() => expect(screen.getByText('10')).toBeInTheDocument())
  })

  it('shows Edit profile button for owner', async () => {
    wrap()
    await waitFor(() => expect(screen.getByTestId('edit-profile-btn')).toBeInTheDocument())
  })

  it('shows Follow button for stranger', async () => {
    vi.mocked(socialApi.getProfile).mockResolvedValue({ data: { user: strangerProfile } } as never)
    vi.mocked(useAuthStore).mockReturnValue({ user: { id: 'u1', username: 'alice', role: 'user' } } as never)
    wrap('bob')
    await waitFor(() => expect(screen.getByTestId('follow-btn')).toBeInTheDocument())
  })

  it('calls follow when follow button clicked', async () => {
    vi.mocked(socialApi.getProfile).mockResolvedValue({ data: { user: strangerProfile } } as never)
    vi.mocked(socialApi.follow).mockResolvedValue({} as never)
    wrap('bob')
    await waitFor(() => screen.getByTestId('follow-btn'))
    await userEvent.click(screen.getByTestId('follow-btn'))
    expect(socialApi.follow).toHaveBeenCalledWith('u2')
  })

  it('enters edit mode on Edit profile click', async () => {
    wrap()
    await waitFor(() => screen.getByTestId('edit-profile-btn'))
    await userEvent.click(screen.getByTestId('edit-profile-btn'))
    expect(screen.getByTestId('edit-username')).toBeInTheDocument()
  })

  it('calls updateProfile on save', async () => {
    vi.mocked(socialApi.updateProfile).mockResolvedValue({ data: { user: aliceProfile } } as never)
    wrap()
    await waitFor(() => screen.getByTestId('edit-profile-btn'))
    await userEvent.click(screen.getByTestId('edit-profile-btn'))
    await userEvent.click(screen.getByTestId('save-profile-btn'))
    expect(socialApi.updateProfile).toHaveBeenCalled()
  })

  it('renders avatar image when avatar_url is set', async () => {
    vi.mocked(socialApi.getProfile).mockResolvedValue({
      data: { user: { ...aliceProfile, avatar_url: 'http://cdn/avatar.jpg' } },
    } as never)
    wrap()
    await waitFor(() => expect(screen.getByRole('img', { name: 'alice' })).toHaveAttribute('src', 'http://cdn/avatar.jpg'))
  })

  it('renders favorite quote when set', async () => {
    vi.mocked(socialApi.getProfile).mockResolvedValue({
      data: { user: { ...aliceProfile, favorite_quote: 'Code is poetry' } },
    } as never)
    wrap()
    await waitFor(() => expect(screen.getByText(/"Code is poetry"/)).toBeInTheDocument())
  })

  it('shows unfollow button for friend relation', async () => {
    vi.mocked(socialApi.getProfile).mockResolvedValue({
      data: { user: { ...strangerProfile, viewer_relation: 'friend' } },
    } as never)
    wrap('bob')
    await waitFor(() => expect(screen.getByTestId('unfollow-btn')).toBeInTheDocument())
  })

  it('calls unfollow when unfollow button is clicked', async () => {
    vi.mocked(socialApi.getProfile).mockResolvedValue({
      data: { user: { ...strangerProfile, viewer_relation: 'friend' } },
    } as never)
    vi.mocked(socialApi.unfollow).mockResolvedValue({} as never)
    wrap('bob')
    await waitFor(() => screen.getByTestId('unfollow-btn'))
    await userEvent.click(screen.getByTestId('unfollow-btn'))
    expect(socialApi.unfollow).toHaveBeenCalledWith('u2')
  })

  it('shows user not found when profile data is missing', async () => {
    vi.mocked(socialApi.getProfile).mockRejectedValue({ response: { status: 404 } })
    wrap('nobody')
    await waitFor(() => expect(screen.getByText(/user not found/i)).toBeInTheDocument())
  })

  it('cancels edit mode without saving', async () => {
    wrap()
    await waitFor(() => screen.getByTestId('edit-profile-btn'))
    await userEvent.click(screen.getByTestId('edit-profile-btn'))
    await userEvent.click(screen.getByRole('button', { name: /cancel/i }))
    expect(screen.queryByTestId('edit-username')).not.toBeInTheDocument()
    expect(screen.getByTestId('profile-username')).toBeInTheDocument()
  })
})
