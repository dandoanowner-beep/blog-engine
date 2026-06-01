import { Link } from 'react-router-dom'
import type { BlogCard as BlogCardType } from '../types'

interface Props {
  blog: BlogCardType
}

export default function BlogCard({ blog }: Props) {
  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-100 overflow-hidden hover:shadow-md transition-shadow" data-testid="blog-card">
      {blog.thumbnail_url && (
        <img
          src={blog.thumbnail_url}
          alt={blog.title}
          className="w-full h-48 object-cover"
        />
      )}
      <div className="p-4">
        <div className="flex items-center gap-2 mb-2">
          {blog.author.avatar_url && (
            <img src={blog.author.avatar_url} alt={blog.author.username} className="w-6 h-6 rounded-full" />
          )}
          <Link to={`/profile/${blog.author.username}`} className="text-sm text-gray-600 hover:text-blue-600">
            {blog.author.username}
          </Link>
          <span className="text-gray-400 text-xs">·</span>
          <span className="text-xs text-gray-400">{blog.read_time_min} min read</span>
        </div>

        <Link to={`/blog/${blog.id}`}>
          <h2 className="font-semibold text-gray-900 mb-1 line-clamp-2 hover:text-blue-600">
            {blog.title}
          </h2>
          <p className="text-gray-500 text-sm line-clamp-2 mb-3">{blog.excerpt}</p>
        </Link>

        <div className="flex flex-wrap gap-1 mb-3">
          {blog.tags.map((tag) => (
            <span key={tag.id} className="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded-full">
              {tag.name}
            </span>
          ))}
        </div>

        <div className="flex items-center gap-4 text-xs text-gray-500">
          <span>👍 {blog.like_count}</span>
          <span>👎 {blog.dislike_count}</span>
          <span>💬 {blog.comment_count}</span>
          {blog.partial && (
            <span className="ml-auto text-blue-500 font-medium">Sign in to read more →</span>
          )}
        </div>
      </div>
    </div>
  )
}
