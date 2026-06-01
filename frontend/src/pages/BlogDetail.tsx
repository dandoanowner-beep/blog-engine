import { useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { blogsApi } from '../api/blogs'
import { useAuthStore } from '../store/auth'
import GuestPrompt from '../components/GuestPrompt'

export default function BlogDetail() {
  const { id } = useParams<{ id: string }>()
  const { user } = useAuthStore()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const [comment, setComment] = useState('')

  const { data: blog, isLoading } = useQuery({
    queryKey: ['blog', id],
    queryFn: () => blogsApi.getBlog(id!).then((r) => r.data),
    enabled: !!id,
  })

  const { data: commentsData } = useQuery({
    queryKey: ['comments', id],
    queryFn: () => blogsApi.getComments(id!, 1).then((r) => r.data),
    enabled: !!id && !!user,
  })

  const reactMutation = useMutation({
    mutationFn: (type: 'like' | 'dislike') => blogsApi.react(id!, type),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['blog', id] }),
  })

  const commentMutation = useMutation({
    mutationFn: (content: string) => blogsApi.createComment(id!, content),
    onSuccess: () => {
      setComment('')
      qc.invalidateQueries({ queryKey: ['comments', id] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: () => blogsApi.deleteBlog(id!),
    onSuccess: () => navigate('/'),
  })

  if (isLoading) {
    return <div className="animate-pulse h-96 bg-gray-100 rounded-lg" />
  }
  if (!blog) {
    return <div className="text-center py-12 text-gray-500">Post not found.</div>
  }

  const isAuthor = user?.id === blog.author.id
  const isMod = user && ['moderator', 'admin', 'owner'].includes(user.role)

  return (
    <div className="max-w-3xl mx-auto">
      {blog.thumbnail_url && (
        <img src={blog.thumbnail_url} alt={blog.title} className="w-full h-64 object-cover rounded-lg mb-6" />
      )}

      <h1 className="text-3xl font-bold mb-3" data-testid="blog-title">{blog.title}</h1>

      <div className="flex items-center gap-3 text-sm text-gray-500 mb-6">
        <Link to={`/profile/${blog.author.username}`} className="font-medium text-gray-700 hover:text-blue-600">
          {blog.author.username}
        </Link>
        <span>·</span>
        <span>{blog.read_time_min} min read</span>
        <span>·</span>
        <span>{new Date(blog.published_at).toLocaleDateString()}</span>
      </div>

      <div
        className="prose max-w-none mb-8"
        dangerouslySetInnerHTML={{ __html: blog.content }}
        data-testid="blog-content"
      />

      {blog.partial && <GuestPrompt />}

      {user && !blog.partial && (
        <>
          <div className="flex items-center gap-4 mb-8">
            <button
              onClick={() => reactMutation.mutate('like')}
              className="flex items-center gap-1 text-sm text-gray-600 hover:text-blue-600"
              aria-label="Like"
            >
              👍 {blog.like_count}
            </button>
            <button
              onClick={() => reactMutation.mutate('dislike')}
              className="flex items-center gap-1 text-sm text-gray-600 hover:text-red-500"
              aria-label="Dislike"
            >
              👎 {blog.dislike_count}
            </button>
            <button onClick={() => blogsApi.share(blog.id, 'facebook')} className="text-sm text-gray-400 hover:text-blue-700">
              Facebook
            </button>
            <button onClick={() => blogsApi.copyLink(blog.id)} className="text-sm text-gray-400 hover:text-gray-700">
              Copy link
            </button>
            {(isAuthor || isMod) && (
              <button
                onClick={() => { if (confirm('Delete this post?')) deleteMutation.mutate() }}
                className="ml-auto text-sm text-red-400 hover:text-red-600"
                data-testid="delete-btn"
              >
                Delete
              </button>
            )}
            {isAuthor && (
              <Link to={`/editor/${blog.id}`} className="text-sm text-gray-400 hover:text-gray-700">
                Edit
              </Link>
            )}
          </div>

          <div className="border-t pt-6">
            <h3 className="font-semibold mb-4">Comments ({commentsData?.items?.length ?? 0})</h3>
            <form
              onSubmit={(e) => { e.preventDefault(); commentMutation.mutate(comment) }}
              className="flex gap-2 mb-6"
            >
              <input
                value={comment}
                onChange={(e) => setComment(e.target.value)}
                placeholder="Add a comment…"
                className="flex-1 border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                data-testid="comment-input"
              />
              <button
                type="submit"
                disabled={!comment.trim() || commentMutation.isPending}
                className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700 disabled:opacity-60"
              >
                Post
              </button>
            </form>
          </div>
        </>
      )}
    </div>
  )
}
