import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { blogsApi } from '../api/blogs'
import { useAuthStore } from '../store/auth'
import BlogCard from '../components/BlogCard'
import Pagination from '../components/Pagination'

type Tab = 'explore' | 'following'

export default function Feed() {
  const { user } = useAuthStore()
  const [tab, setTab] = useState<Tab>('explore')
  const [page, setPage] = useState(1)

  const exploreQuery = useQuery({
    queryKey: ['feed', 'explore', page],
    queryFn: () => blogsApi.getExploreFeed(page).then((r) => r.data),
    enabled: tab === 'explore',
  })

  const followingQuery = useQuery({
    queryKey: ['feed', 'following', page],
    queryFn: () => blogsApi.getFollowingFeed(page).then((r) => r.data),
    enabled: tab === 'following' && !!user,
  })

  const active = tab === 'explore' ? exploreQuery : followingQuery
  const blogs = active.data?.blogs ?? []
  const total = active.data?.total ?? 0

  return (
    <div>
      <div className="flex gap-4 mb-6 border-b border-gray-200">
        {(['explore', 'following'] as Tab[]).map((t) => (
          <button
            key={t}
            onClick={() => { setTab(t); setPage(1) }}
            className={`pb-2 text-sm font-medium capitalize border-b-2 transition-colors ${
              tab === t ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-800'
            }`}
            data-testid={`tab-${t}`}
          >
            {t === 'following' ? 'Following' : 'Explore'}
          </button>
        ))}
      </div>

      {active.isLoading && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="h-64 bg-gray-100 rounded-lg animate-pulse" />
          ))}
        </div>
      )}

      {!active.isLoading && tab === 'following' && !user && (
        <div className="text-center py-12 text-gray-500">
          <p className="mb-2">Sign in to see posts from people you follow.</p>
        </div>
      )}

      {!active.isLoading && blogs.length === 0 && (tab === 'explore' || user) && (
        <div className="text-center py-12 text-gray-400">No posts yet.</div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {blogs.map((blog) => (
          <BlogCard key={blog.id} blog={blog} />
        ))}
      </div>

      <Pagination page={page} total={total} perPage={9} onPageChange={setPage} />
    </div>
  )
}
