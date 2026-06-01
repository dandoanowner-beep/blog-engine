import { useState, type FormEvent } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link, useSearchParams } from 'react-router-dom'
import { socialApi } from '../api/social'
import BlogCard from '../components/BlogCard'

export default function Search() {
  const [params, setParams] = useSearchParams()
  const [input, setInput] = useState(params.get('q') ?? '')
  const q = params.get('q') ?? ''

  const { data, isLoading } = useQuery({
    queryKey: ['search', q],
    queryFn: () => socialApi.search(q, 1).then((r) => r.data),
    enabled: q.length >= 2,
  })

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    if (input.trim()) setParams({ q: input.trim() })
  }

  return (
    <div>
      <form onSubmit={handleSubmit} className="flex gap-2 mb-6">
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Search blogs, users, tags…"
          className="flex-1 border border-gray-300 rounded px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
          data-testid="search-input"
        />
        <button type="submit" className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">
          Search
        </button>
      </form>

      {isLoading && <div className="text-gray-400 text-sm">Searching…</div>}

      {data && (
        <div className="space-y-8">
          {data.blogs.total > 0 && (
            <section>
              <h2 className="font-semibold mb-3">Posts ({data.blogs.total})</h2>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {data.blogs.items.map((b) => <BlogCard key={b.id} blog={b} />)}
              </div>
            </section>
          )}

          {data.users.total > 0 && (
            <section>
              <h2 className="font-semibold mb-3">People ({data.users.total})</h2>
              <div className="space-y-2">
                {data.users.items.map((u) => (
                  <Link key={u.id} to={`/profile/${u.username}`} className="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 border border-gray-100">
                    {u.avatar_url
                      ? <img src={u.avatar_url} alt={u.username} className="w-8 h-8 rounded-full" />
                      : <div className="w-8 h-8 rounded-full bg-blue-100 text-blue-600 flex items-center justify-center text-sm font-medium">
                          {u.username[0].toUpperCase()}
                        </div>
                    }
                    <div>
                      <p className="font-medium text-sm">{u.username}</p>
                      {u.bio && <p className="text-xs text-gray-500 truncate max-w-xs">{u.bio}</p>}
                    </div>
                    <span className="ml-auto text-xs text-gray-400">{u.follower_count} followers</span>
                  </Link>
                ))}
              </div>
            </section>
          )}

          {data.tags.total > 0 && (
            <section>
              <h2 className="font-semibold mb-3">Tags ({data.tags.total})</h2>
              <div className="flex flex-wrap gap-2">
                {data.tags.items.map((tag) => (
                  <span key={tag.id} className="bg-gray-100 text-gray-700 px-3 py-1 rounded-full text-sm">
                    #{tag.name}
                  </span>
                ))}
              </div>
            </section>
          )}

          {data.blogs.total === 0 && data.users.total === 0 && data.tags.total === 0 && (
            <div className="text-center py-12 text-gray-400">No results for "{q}"</div>
          )}
        </div>
      )}
    </div>
  )
}
