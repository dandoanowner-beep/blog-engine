import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { blogsApi } from '../api/blogs'
import BlogCard from '../components/BlogCard'
import Pagination from '../components/Pagination'

// CR-001 personal-blog pivot: one article feed (the owner's published posts).
// The Explore/Following tab pair was removed with the multi-writer model.
export default function Feed() {
  const [page, setPage] = useState(1)

  const { data, isLoading } = useQuery({
    queryKey: ['feed', 'articles', page],
    queryFn: () => blogsApi.getArticlesFeed(page).then((r) => r.data),
  })

  const blogs = data?.blogs ?? []
  const total = data?.total ?? 0

  return (
    <div>
      {isLoading && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="h-64 bg-gray-100 rounded-lg animate-pulse" />
          ))}
        </div>
      )}

      {!isLoading && blogs.length === 0 && (
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
