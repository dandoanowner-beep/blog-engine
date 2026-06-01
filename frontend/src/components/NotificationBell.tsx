import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { socialApi } from '../api/social'

export default function NotificationBell() {
  const [open, setOpen] = useState(false)
  const qc = useQueryClient()

  const { data } = useQuery({
    queryKey: ['notifications'],
    queryFn: () => socialApi.getNotifications(1).then((r) => r.data),
    refetchInterval: 30000,
  })

  const markAll = useMutation({
    mutationFn: () => socialApi.markAllRead(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['notifications'] }),
  })

  const unread = data?.unread_count ?? 0

  return (
    <div className="relative">
      <button
        onClick={() => setOpen((o) => !o)}
        className="relative p-1"
        aria-label="Notifications"
        data-testid="notification-bell"
      >
        🔔
        {unread > 0 && (
          <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center">
            {unread > 9 ? '9+' : unread}
          </span>
        )}
      </button>

      {open && (
        <div className="absolute right-0 mt-2 w-80 bg-white shadow-lg rounded-lg border border-gray-100 z-20">
          <div className="flex items-center justify-between px-4 py-2 border-b">
            <span className="font-medium text-sm">Notifications</span>
            <button onClick={() => markAll.mutate()} className="text-xs text-blue-600 hover:underline">
              Mark all read
            </button>
          </div>
          <ul className="max-h-64 overflow-y-auto divide-y" data-testid="notification-list">
            {data?.notifications.length === 0 && (
              <li className="px-4 py-3 text-sm text-gray-500">No notifications</li>
            )}
            {data?.notifications.map((n) => (
              <li key={n.id} className={`px-4 py-3 text-sm ${!n.read ? 'bg-blue-50' : ''}`}>
                <span className="font-medium">{n.actor?.username}</span>{' '}
                {n.type.replace(/_/g, ' ')}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}
