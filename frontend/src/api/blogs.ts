import api from './client'
import type { Blog, BlogCard } from '../types'

export interface CreateBlogInput {
  title: string
  content: string
  thumbnail_url?: string
  privacy: 'public' | 'friend_only' | 'only_me'
  status: 'draft' | 'published'
  tag_names: string[]
  category_ids: string[]
}

export const blogsApi = {
  getExploreFeed: (page = 1, tag?: string, category?: string) =>
    api.get<{ blogs: BlogCard[]; total: number; page: number; per_page: number }>(
      '/blogs/feed/explore',
      { params: { page, tag, category } }
    ),

  getFollowingFeed: (page = 1) =>
    api.get<{ blogs: BlogCard[]; total: number; page: number }>(
      '/blogs/feed/following',
      { params: { page } }
    ),

  getBlog: (id: string) =>
    api.get<Blog & { partial?: boolean }>(`/blogs/${id}`),

  createBlog: (input: CreateBlogInput) =>
    api.post<Blog>('/blogs', input),

  updateBlog: (id: string, input: Partial<CreateBlogInput>) =>
    api.patch<Blog>(`/blogs/${id}`, input),

  deleteBlog: (id: string) =>
    api.delete(`/blogs/${id}`),

  react: (blogId: string, type: 'like' | 'dislike') =>
    api.post<{ like_count: number; dislike_count: number }>(`/blogs/${blogId}/react`, { type }),

  removeReaction: (blogId: string) =>
    api.delete<{ like_count: number; dislike_count: number }>(`/blogs/${blogId}/react`),

  getComments: (blogId: string, page = 1) =>
    api.get(`/blogs/${blogId}/comments`, { params: { page } }),

  createComment: (blogId: string, content: string, parentId?: string) =>
    api.post(`/blogs/${blogId}/comments`, { content, parent_id: parentId }),

  deleteComment: (commentId: string) =>
    api.delete(`/comments/${commentId}`),

  uploadImage: (file: File) => {
    const form = new FormData()
    form.append('file', file)
    return api.post<{ url: string }>('/uploads/image', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },

  share: (blogId: string, platform: 'facebook' | 'zalo') => {
    const url = `${window.location.origin}/blog/${blogId}`
    if (platform === 'facebook') {
      window.open(`https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(url)}`, '_blank')
    } else {
      window.open(`https://zalo.me/share?url=${encodeURIComponent(url)}`, '_blank')
    }
  },

  copyLink: (blogId: string) => {
    const url = `${window.location.origin}/blog/${blogId}`
    navigator.clipboard.writeText(url)
  },
}
