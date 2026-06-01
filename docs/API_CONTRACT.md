# API Contract — Blog Engine
# Version: 1.0 — 2026-05-30
# Base URL: /api/v1
# Auth: Bearer JWT in Authorization header (or httpOnly cookie)

---

## Auth

### POST /auth/register
Request: `{ email, username, password }`
Response 201: `{ user_id, message: "Verification email sent" }`
Response 400: `{ error: "Email already in use" | "Password too short" }`

### POST /auth/login
Request: `{ email, password }`
Response 200: `{ access_token, user: { id, username, role } }` + sets httpOnly refresh cookie
Response 401: `{ error: "Invalid email or password" }`
Response 423: `{ error: "Account locked", locked_until }`

### POST /auth/refresh
Request: refresh token cookie
Response 200: `{ access_token }`
Response 401: `{ error: "Invalid or expired refresh token" }`

### GET /auth/google
Response 302: redirect to Google OAuth

### GET /auth/google/callback
Response 302: redirect to frontend with access token

### GET /auth/verify?token=
Response 200: `{ message: "Email verified" }`
Response 400: `{ error: "Invalid or expired token" }`

### POST /auth/resend-verification
Auth: required
Response 200: `{ message: "Verification email sent" }`
Response 429: `{ error: "Please wait 5 minutes before requesting again" }`

### POST /auth/forgot-password
Request: `{ email }`
Response 200: `{ message: "Reset email sent if account exists" }`

### POST /auth/reset-password
Request: `{ token, password }`
Response 200: `{ message: "Password updated" }`
Response 400: `{ error: "Invalid or expired token" }`

### POST /auth/logout
Auth: required
Response 200: clears cookies

---

## Blogs

### POST /blogs
Auth: required + verified
Request: `{ title, content, thumbnail_url, privacy, tag_ids[], category_ids[], status }`
Response 201: `{ blog }` (full blog object)
Response 403: `{ error: "Email not verified" }`
Response 422: `{ error: "Title required" | "At least one tag required" }`

### GET /blogs/:id
Auth: optional
Response 200: `{ blog }` (full content if auth + permitted; partial if guest)
Response 403: `{ error: "Access denied", hint: "Sign in to read" }`
Response 404: `{ error: "Not found" }`

### PATCH /blogs/:id
Auth: required (author only)
Request: `{ title?, content?, thumbnail_url?, privacy?, tag_ids[]?, category_ids[]?, status? }`
Response 200: `{ blog }`
Response 403: `{ error: "Not your blog" }`

### DELETE /blogs/:id
Auth: required (author or moderator+)
Response 204: no content
Response 403: `{ error: "Forbidden" }`

### GET /blogs/feed/explore?page=&tag=&category=
Auth: optional
Response 200: `{ blogs: [BlogCard], total, page, per_page: 12 }`

### GET /blogs/feed/following?page=
Auth: required
Response 200: `{ blogs: [BlogCard], total, page, per_page: 12 }`

---

## Uploads

### POST /uploads/image
Auth: required
Request: multipart/form-data `{ file, blog_id? }`
Response 201: `{ url: "https://pub-xxx.r2.dev/<key>", r2_key: "<key>" }` (Cloudflare R2 public URL)
Response 413: `{ error: "File exceeds 5MB" }`
Response 415: `{ error: "Only JPEG, PNG, WEBP allowed" }`

---

## Tags & Categories

### GET /tags?q=
Response 200: `{ tags: [{ id, name, slug }] }`

### POST /tags
Auth: required
Request: `{ name }`
Response 201: `{ tag }`

### GET /categories
Response 200: `{ categories: [{ id, name, slug, created_by }] }`

### POST /categories
Auth: required
Request: `{ name }`
Response 201: `{ category }`

---

## Social — Follows

### POST /users/:id/follow
Auth: required
Response 201: `{ message: "Following" }`
Response 409: `{ error: "Already following" }`

### DELETE /users/:id/follow
Auth: required
Response 204: no content

---

## Social — Friends

### POST /users/:id/friend-request
Auth: required
Response 201: `{ request_id, status: "pending" }`
Response 409: `{ error: "Request already pending" }`

### PATCH /friend-requests/:id
Auth: required (receiver only)
Request: `{ action: "accept" | "reject" }`
Response 200: `{ status }`

### DELETE /users/:id/friend
Auth: required
Response 204: no content

---

## Social — Reactions

### POST /blogs/:id/react
Auth: required
Request: `{ type: "like" | "dislike" }`
Response 200: `{ like_count, dislike_count, user_reaction }`

### DELETE /blogs/:id/react
Auth: required
Response 200: `{ like_count, dislike_count, user_reaction: null }`

---

## Social — Comments

### POST /blogs/:id/comments
Auth: required
Request: `{ content, parent_id? }`
Response 201: `{ comment }`

### GET /blogs/:id/comments?page=
Auth: optional
Response 200: `{ comments: [CommentTree], total, page }`

### DELETE /comments/:id
Auth: required (author or moderator+)
Response 204: no content

---

## Social — Reports

### POST /reports
Auth: required
Request: `{ blog_id? | comment_id?, reason }`
Reason enum: `spam | inappropriate | harassment | misinformation | other`
Response 201: `{ message: "Report submitted" }`
Response 409: `{ error: "Already reported" }`

---

## Social — Block

### POST /users/:id/block
Auth: required
Response 201: `{ message: "User blocked" }`

### DELETE /users/:id/block
Auth: required
Response 204: no content

---

## Notifications

### GET /notifications?page=
Auth: required
Response 200: `{ notifications: [Notification], unread_count, total, page }`

### PATCH /notifications/:id/read
Auth: required
Response 200: `{ read: true }`

### PATCH /notifications/read-all
Auth: required
Response 200: `{ message: "All marked as read" }`

---

## Search

### GET /search?q=&page=
Auth: optional
Response 200:
```json
{
  "query": "go tutorial",
  "blogs": { "items": [BlogCard], "total": 42, "page": 1 },
  "users": { "items": [UserPreview], "total": 3 },
  "tags":  { "items": [Tag], "total": 7 }
}
```

---

## User Profiles

### GET /users/:username
Auth: optional
Response 200: `{ user: UserProfile }` (blog visibility filtered by relationship)

### PATCH /users/me
Auth: required
Request: `{ username?, bio?, favorite_quote?, avatar_url? }`
Response 200: `{ user }`

---

## Admin Dashboard

### GET /admin/users?page=&role=
Auth: required (admin+)
Response 200: `{ users: [UserAdmin], total, page }`

### PATCH /admin/users/:id/role
Auth: required (admin+)
Request: `{ role: "user" | "moderator" | "admin" }`
Response 200: `{ user }`

### GET /admin/reports?status=pending&page=
Auth: required (moderator+)
Response 200: `{ reports: [Report], total, page }`

### PATCH /admin/reports/:id/resolve
Auth: required (moderator+)
Request: `{ action: "delete_content" | "dismiss" }`
Response 200: `{ report }`

### GET /admin/stats
Auth: required (admin+)
Response 200:
```json
{
  "total_users": 1200,
  "total_blogs": 4500,
  "total_comments": 18000,
  "new_signups_today": 12,
  "new_signups_this_week": 67
}
```

---

## Shared Object Shapes

### BlogCard
```json
{
  "id": "uuid",
  "title": "string",
  "excerpt": "string (100 chars)",
  "thumbnail_url": "string",
  "author": { "id": "uuid", "username": "string", "avatar_url": "string" },
  "read_time_min": 3,
  "tags": [{ "id": "uuid", "name": "string", "slug": "string" }],
  "like_count": 42,
  "dislike_count": 2,
  "comment_count": 15,
  "privacy": "public",
  "published_at": "ISO8601"
}
```

### CommentTree
```json
{
  "id": "uuid",
  "content": "string",
  "author": { "id": "uuid", "username": "string", "avatar_url": "string" },
  "created_at": "ISO8601",
  "replies": [CommentTree]
}
```

### Notification
```json
{
  "id": "uuid",
  "type": "like_blog",
  "actor": { "id": "uuid", "username": "string", "avatar_url": "string" },
  "blog_id": "uuid | null",
  "comment_id": "uuid | null",
  "read": false,
  "created_at": "ISO8601"
}
```
