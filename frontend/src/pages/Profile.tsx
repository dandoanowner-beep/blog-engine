import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { socialApi } from '../api/social'
import { useAuthStore } from '../store/auth'

export default function Profile() {
  const { username } = useParams<{ username: string }>()
  const { user } = useAuthStore()
  const qc = useQueryClient()
  const [editMode, setEditMode] = useState(false)
  const [bio, setBio] = useState('')
  const [quote, setQuote] = useState('')
  const [newUsername, setNewUsername] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['profile', username],
    queryFn: () => socialApi.getProfile(username!).then((r) => r.data.user),
    enabled: !!username,
  })

  useEffect(() => {
    if (!data) return
    setBio(data.bio ?? '')
    setQuote(data.favorite_quote ?? '')
    setNewUsername(data.username)
  }, [data])

  const followMutation = useMutation({
    mutationFn: () => data ? socialApi.follow(data.id) : Promise.reject(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['profile', username] }),
  })

  const unfollowMutation = useMutation({
    mutationFn: () => data ? socialApi.unfollow(data.id) : Promise.reject(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['profile', username] }),
  })

  const updateMutation = useMutation({
    mutationFn: () => socialApi.updateProfile({ bio, favorite_quote: quote, username: newUsername }),
    onSuccess: () => {
      setEditMode(false)
      qc.invalidateQueries({ queryKey: ['profile', username] })
    },
  })

  if (isLoading) return <div className="animate-pulse h-48 bg-gray-100 rounded-lg" />
  if (!data) return <div className="text-center py-12 text-gray-500">User not found.</div>

  const isOwner = data.viewer_relation === 'owner'

  return (
    <div className="max-w-2xl mx-auto">
      <div className="bg-white rounded-lg border border-gray-100 p-6 mb-6">
        <div className="flex items-start gap-4">
          {data.avatar_url
            ? <img src={data.avatar_url} alt={data.username} className="w-20 h-20 rounded-full object-cover" />
            : <div className="w-20 h-20 rounded-full bg-blue-100 text-blue-600 flex items-center justify-center text-2xl font-bold">
                {data.username[0].toUpperCase()}
              </div>
          }

          <div className="flex-1">
            {editMode ? (
              <div className="space-y-2">
                <input
                  value={newUsername}
                  onChange={(e) => setNewUsername(e.target.value)}
                  className="border border-gray-300 rounded px-3 py-1 text-sm w-full"
                  data-testid="edit-username"
                />
                <input
                  value={bio}
                  onChange={(e) => setBio(e.target.value)}
                  placeholder="Bio"
                  className="border border-gray-300 rounded px-3 py-1 text-sm w-full"
                />
                <input
                  value={quote}
                  onChange={(e) => setQuote(e.target.value)}
                  placeholder="Favourite quote"
                  className="border border-gray-300 rounded px-3 py-1 text-sm w-full"
                />
                <div className="flex gap-2">
                  <button
                    onClick={() => updateMutation.mutate()}
                    disabled={updateMutation.isPending}
                    className="bg-blue-600 text-white px-3 py-1 rounded text-sm"
                    data-testid="save-profile-btn"
                  >
                    Save
                  </button>
                  <button onClick={() => setEditMode(false)} className="border border-gray-300 px-3 py-1 rounded text-sm">
                    Cancel
                  </button>
                </div>
              </div>
            ) : (
              <>
                <h1 className="text-xl font-bold" data-testid="profile-username">{data.username}</h1>
                {data.bio && <p className="text-gray-600 text-sm mt-1">{data.bio}</p>}
                {data.favorite_quote && (
                  <p className="text-gray-400 text-sm italic mt-1">"{data.favorite_quote}"</p>
                )}
              </>
            )}

            <div className="flex gap-4 mt-3 text-sm text-gray-500">
              <span><strong>{data.follower_count}</strong> followers</span>
              <span><strong>{data.following_count}</strong> following</span>
              <span><strong>{data.friend_count}</strong> friends</span>
            </div>
          </div>

          <div className="flex flex-col gap-2">
            {isOwner && !editMode && (
              <button
                onClick={() => setEditMode(true)}
                className="border border-gray-300 px-3 py-1 rounded text-sm hover:bg-gray-50"
                data-testid="edit-profile-btn"
              >
                Edit profile
              </button>
            )}
            {user && !isOwner && data.viewer_relation === 'stranger' && (
              <button
                onClick={() => followMutation.mutate()}
                disabled={followMutation.isPending}
                className="bg-blue-600 text-white px-3 py-1 rounded text-sm hover:bg-blue-700"
                data-testid="follow-btn"
              >
                Follow
              </button>
            )}
            {user && !isOwner && data.viewer_relation === 'friend' && (
              <button
                onClick={() => unfollowMutation.mutate()}
                className="border border-gray-300 px-3 py-1 rounded text-sm hover:bg-gray-50"
                data-testid="unfollow-btn"
              >
                Unfollow
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
