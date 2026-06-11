import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { blogsApi } from '../api/blogs'
import BlogCard from '../components/BlogCard'
import Pagination from '../components/Pagination'

// FR-CR002-003: browse categories → click one to see its articles.
export default function Categories() {
  const [selected, setSelected] = useState<string | null>(null)
  const [page, setPage] = useState(1)

  const categoriesQuery = useQuery({
    queryKey: ['categories'],
    queryFn: () => blogsApi.getCategories().then((r) => r.data),
  })

  const articlesQuery = useQuery({
    queryKey: ['feed', 'category', selected, page],
    queryFn: () => blogsApi.getArticlesFeed(page, selected ?? undefined).then((r) => r.data),
    enabled: selected !== null,
  })

  const categories = categoriesQuery.data?.categories ?? []
  const blogs = articlesQuery.data?.blogs ?? []
  const total = articlesQuery.data?.total ?? 0

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Categories</h1>

      {categoriesQuery.isLoading && (
        <div className="flex gap-2 flex-wrap">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="h-9 w-28 bg-gray-100 rounded-full animate-pulse" />
          ))}
        </div>
      )}

      {!categoriesQuery.isLoading && categories.length === 0 && (
        <div className="text-center py-12 text-gray-400">No categories yet.</div>
      )}

      <div className="flex gap-2 flex-wrap mb-8">
        {categories.map((c) => (
          <button
            key={c.id}
            onClick={() => { setSelected(c.slug); setPage(1) }}
            className={`flex items-center gap-2 px-4 py-2 rounded-full border text-sm transition-colors ${
              selected === c.slug
                ? 'bg-blue-600 text-white border-blue-600'
                : 'bg-white text-gray-700 border-gray-300 hover:border-blue-400'
            }`}
            data-testid={`category-${c.slug}`}
          >
            {c.name}
            <span className={`text-xs px-2 py-0.5 rounded-full ${
              selected === c.slug ? 'bg-blue-500' : 'bg-gray-100 text-gray-500'
            }`}>
              {c.blog_count}
            </span>
          </button>
        ))}
      </div>

      {selected && (
        <>
          {articlesQuery.isLoading && (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {Array.from({ length: 3 }).map((_, i) => (
                <div key={i} className="h-64 bg-gray-100 rounded-lg animate-pulse" />
              ))}
            </div>
          )}

          {!articlesQuery.isLoading && blogs.length === 0 && (
            <div className="text-center py-12 text-gray-400">No articles in this category yet.</div>
          )}

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {blogs.map((blog) => (
              <BlogCard key={blog.id} blog={blog} />
            ))}
          </div>

          <Pagination page={page} total={total} perPage={9} onPageChange={setPage} />
        </>
      )}
    </div>
  )
}
