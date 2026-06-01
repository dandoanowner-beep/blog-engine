import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import api from '../api/client'
import type { User, UserProfile } from '../types'

interface AdminStats {
  total_users: number
  total_blogs: number
  pending_reports: number
}

interface Report {
  id: string
  reason: string
  reporter: { username: string }
  blog_id?: string
  comment_id?: string
  created_at: string
}

export default function Admin() {
  const [tab, setTab] = useState<'stats' | 'users' | 'reports'>('stats')

  const statsQuery = useQuery({
    queryKey: ['admin', 'stats'],
    queryFn: () => api.get<AdminStats>('/admin/stats').then((r) => r.data),
  })

  const usersQuery = useQuery({
    queryKey: ['admin', 'users'],
    queryFn: () => api.get<{ users: (User & UserProfile)[] }>('/admin/users').then((r) => r.data.users),
    enabled: tab === 'users',
  })

  const reportsQuery = useQuery({
    queryKey: ['admin', 'reports'],
    queryFn: () => api.get<{ reports: Report[] }>('/admin/reports').then((r) => r.data.reports),
    enabled: tab === 'reports',
  })

  const promoteUser = async (userId: string, role: string) => {
    await api.patch(`/admin/users/${userId}`, { role })
    usersQuery.refetch()
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Admin Dashboard</h1>

      <div className="flex gap-4 mb-6 border-b border-gray-200">
        {(['stats', 'users', 'reports'] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`pb-2 text-sm font-medium capitalize border-b-2 transition-colors ${
              tab === t ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-800'
            }`}
            data-testid={`admin-tab-${t}`}
          >
            {t}
          </button>
        ))}
      </div>

      {tab === 'stats' && statsQuery.data && (
        <div className="grid grid-cols-3 gap-4">
          {[
            { label: 'Total Users', value: statsQuery.data.total_users },
            { label: 'Total Posts', value: statsQuery.data.total_blogs },
            { label: 'Pending Reports', value: statsQuery.data.pending_reports },
          ].map(({ label, value }) => (
            <div key={label} className="bg-white rounded-lg border border-gray-100 p-6 text-center">
              <p className="text-3xl font-bold text-blue-600" data-testid={`stat-${label.toLowerCase().replace(/ /g, '-')}`}>{value}</p>
              <p className="text-sm text-gray-500 mt-1">{label}</p>
            </div>
          ))}
        </div>
      )}

      {tab === 'users' && (
        <div className="bg-white rounded-lg border border-gray-100 overflow-hidden">
          {usersQuery.isLoading && <div className="p-4 text-gray-400">Loading users…</div>}
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-100">
              <tr>
                <th className="text-left px-4 py-2">User</th>
                <th className="text-left px-4 py-2">Role</th>
                <th className="text-left px-4 py-2">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {usersQuery.data?.map((u) => (
                <tr key={u.id} data-testid="user-row">
                  <td className="px-4 py-2 font-medium">{u.username}</td>
                  <td className="px-4 py-2 text-gray-500 capitalize">{u.role}</td>
                  <td className="px-4 py-2">
                    {u.role === 'user' && (
                      <button
                        onClick={() => promoteUser(u.id, 'moderator')}
                        className="text-xs text-blue-600 hover:underline"
                        data-testid="promote-btn"
                      >
                        Promote to Mod
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {tab === 'reports' && (
        <div className="space-y-3">
          {reportsQuery.isLoading && <div className="text-gray-400">Loading reports…</div>}
          {reportsQuery.data?.length === 0 && (
            <div className="text-center py-12 text-gray-400">No pending reports</div>
          )}
          {reportsQuery.data?.map((r) => (
            <div key={r.id} className="bg-white rounded-lg border border-gray-100 p-4" data-testid="report-row">
              <div className="flex items-center justify-between">
                <div>
                  <span className="font-medium text-sm">{r.reporter.username}</span>
                  <span className="text-gray-500 text-sm"> reported {r.blog_id ? 'a post' : 'a comment'}</span>
                </div>
                <span className="text-xs text-gray-400">{new Date(r.created_at).toLocaleDateString()}</span>
              </div>
              <p className="text-sm text-gray-600 mt-1 italic">"{r.reason}"</p>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
