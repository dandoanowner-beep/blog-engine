// Core domain types matching API_CONTRACT.md

export interface User {
  id: string
  username: string
  email?: string
  avatar_url?: string
  role: 'guest' | 'user' | 'moderator' | 'admin' | 'owner'
  verified: boolean
}

export interface BlogCard {
  id: string
  title: string
  excerpt: string
  thumbnail_url?: string
  author: { id: string; username: string; avatar_url?: string }
  read_time_min: number
  tags: { id: string; name: string; slug: string }[]
  like_count: number
  dislike_count: number
  comment_count: number
  privacy: 'public' | 'friend_only' | 'only_me'
  published_at: string
  partial?: boolean
  title_en?: string
  translation_status?: 'none' | 'pending' | 'done' | 'failed'
}

export interface Blog extends BlogCard {
  content: string
  body_en?: string
}

export interface Comment {
  id: string
  content: string
  author: { id: string; username: string; avatar_url?: string }
  parent_id?: string
  created_at: string
  replies?: Comment[]
}

export interface Notification {
  id: string
  type: string
  actor?: { id: string; username: string; avatar_url?: string }
  blog_id?: string
  comment_id?: string
  read: boolean
  created_at: string
}

export interface UserProfile {
  id: string
  username: string
  bio?: string
  favorite_quote?: string
  avatar_url?: string
  follower_count: number
  following_count: number
  friend_count: number
  viewer_relation: 'owner' | 'friend' | 'stranger' | 'guest'
}

export interface SearchResults {
  query: string
  blogs: { items: BlogCard[]; total: number; page: number }
  users: { items: UserProfile[]; total: number }
  tags: { items: { id: string; name: string; slug: string }[]; total: number }
}

export interface PagedResponse<T> {
  items: T[]
  total: number
  page: number
  per_page: number
}
