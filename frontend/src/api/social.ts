import api from './client'
import type { Notification, UserProfile, SearchResults } from '../types'

export const socialApi = {
  follow: (userId: string) => api.post(`/users/${userId}/follow`),
  unfollow: (userId: string) => api.delete(`/users/${userId}/follow`),

  sendFriendRequest: (userId: string) => api.post(`/users/${userId}/friend-request`),
  respondFriendRequest: (requestId: string, action: 'accept' | 'reject') =>
    api.patch(`/friend-requests/${requestId}`, { action }),
  unfriend: (userId: string) => api.delete(`/users/${userId}/friend`),

  blockUser: (userId: string) => api.post(`/users/${userId}/block`),
  unblockUser: (userId: string) => api.delete(`/users/${userId}/block`),

  report: (payload: { blog_id?: string; comment_id?: string; reason: string }) =>
    api.post('/reports', payload),

  getNotifications: (page = 1) =>
    api.get<{ notifications: Notification[]; unread_count: number; total: number }>(
      '/notifications', { params: { page } }
    ),
  markRead: (id: string) => api.patch(`/notifications/${id}/read`),
  markAllRead: () => api.patch('/notifications/read-all'),

  search: (q: string, page = 1) =>
    api.get<SearchResults>('/search', { params: { q, page } }),

  getProfile: (username: string) =>
    api.get<{ user: UserProfile }>(`/users/${username}`),
  updateProfile: (data: { username?: string; bio?: string; favorite_quote?: string; avatar_url?: string }) =>
    api.patch<{ user: UserProfile }>('/users/me', data),
}
