# Functional Requirements Document (FRD)
# Project: Blog Engine
# Version: 1.0 — 2026-05-30

---

## 1. Overview

A full-featured blog platform where an Owner can publish blogs, and registered users (friends) can sign up, sign in, and publish their own content. The platform supports rich social features including following, friends, comments, reactions, sharing, and notifications. Available on web (React) and mobile (React Native / iOS + Android).

---

## 2. User Roles & Permissions

| Role | Description | Key Permissions |
|------|-------------|-----------------|
| **Guest** | Unauthenticated visitor | Read top portion of public blogs, browse blog cards, cannot comment |
| **User** | Registered & verified member | Full read, publish blogs, comment, like/dislike, follow, friend, share, report |
| **Moderator** | Elevated user appointed by Admin | All User permissions + delete any blog or comment, receive report notifications |
| **Admin** | Platform manager | All Moderator permissions + manage users, promote to Moderator, view admin dashboard |
| **Owner** | Platform owner (you) | Full access to everything including admin dashboard, all permissions |

### Permission Matrix

| Action | Guest | User | Moderator | Admin | Owner |
|--------|-------|------|-----------|-------|-------|
| Read public blogs (partial) | ✓ | ✓ | ✓ | ✓ | ✓ |
| Read public blogs (full) | ✗ | ✓ | ✓ | ✓ | ✓ |
| Read friend-only blogs | ✗ | Friends only | ✓ | ✓ | ✓ |
| Read only-me blogs | ✗ | Owner of blog | ✓ | ✓ | ✓ |
| Publish blogs | ✗ | ✓ (verified) | ✓ | ✓ | ✓ |
| Comment | ✗ | ✓ | ✓ | ✓ | ✓ |
| Like/Dislike | ✗ | ✓ | ✓ | ✓ | ✓ |
| Follow users | ✗ | ✓ | ✓ | ✓ | ✓ |
| Send friend requests | ✗ | ✓ | ✓ | ✓ | ✓ |
| Report blog/comment | ✗ | ✓ | ✓ | ✓ | ✓ |
| Delete own blog | ✗ | ✓ | ✓ | ✓ | ✓ |
| Delete any blog | ✗ | ✗ | ✓ | ✓ | ✓ |
| Delete any comment | ✗ | ✗ | ✓ | ✓ | ✓ |
| Promote to Moderator | ✗ | ✗ | ✗ | ✓ | ✓ |
| Admin dashboard | ✗ | ✗ | ✗ | ✓ | ✓ |
| Block users | ✗ | ✓ | ✓ | ✓ | ✓ |

---

## 3. Authentication & Account Management

### FR-AUTH-001: Email + Password Registration
- User provides email, username, password
- System sends verification email
- User must verify email before publishing blogs
- User can read and interact (except publish) before verification

### FR-AUTH-002: Google OAuth Registration / Login
- One-click sign up or sign in with Google account
- Google-authenticated users are considered verified automatically

### FR-AUTH-003: Email Verification
- Verification link sent on registration
- Link expires after 24 hours
- User can request a new verification email
- Unverified users cannot publish blogs

### FR-AUTH-004: Login
- Email + password login
- Google OAuth login
- JWT-based session management
- Refresh token support

### FR-AUTH-005: Password Reset
- User requests reset via email
- System sends reset link (expires in 1 hour)
- User sets new password via link

### FR-AUTH-006: Account Restrictions
- Users CANNOT delete their account
- Users CAN delete their own published blogs
- On blog deletion: blog, all comments, and all reactions on that blog are permanently removed

### FR-AUTH-007: Block System
- User A can block User B
- After block: neither sees the other's blogs in any feed
- Blocked user cannot view blocker's profile or any of their content
- Block is silent (blocked user is not notified)

---

## 4. Blog Management

### FR-BLOG-001: Blog Creation
- Rich WYSIWYG editor (formatting toolbar: bold, italic, underline, heading levels, lists, blockquote, divider)
- Markdown support
- Code block support with syntax highlighting
- One thumbnail/cover image (max 5MB, uploaded from device) — displayed on blog card
- Multiple inline images in body (max 5MB each, uploaded from device)
- At least one tag required (user-defined)
- At least one category required (predefined or user-created)
- Privacy mode selection: **Public** / **Friend-only** / **Only-me**

### FR-BLOG-002: Draft System
- Writer can save blog as draft at any time
- Drafts are private (Only-me) until published
- Writer can edit and publish a draft later
- Drafts appear in writer's own dashboard / profile

### FR-BLOG-003: Blog Privacy Modes
| Mode | Visible to |
|------|-----------|
| **Public** | Everyone (including guests, partial for guests) |
| **Friend-only** | Only mutual friends of the writer |
| **Only-me** | Only the writer themselves |

### FR-BLOG-004: Blog Editing & Deletion
- Writer can edit a published blog at any time
- Writer can delete their own blog (removes blog + comments + reactions)
- Moderator/Admin/Owner can delete any blog

### FR-BLOG-005: Tags & Categories
- Tags: free-form, user-defined keywords (e.g. "javascript", "travel-tips")
- Categories: predefined base list (Owner/Admin creates) + users can create custom categories
- Users can browse/filter feed by tag or category

### FR-BLOG-006: Guest Partial Read
- Guest opens a public blog card
- Reads the top portion (approx. first 30% of content)
- Lower content is blurred/hidden
- Prompt displayed: "Sign up or log in to read the full blog"

---

## 5. Feed & Discovery

### FR-FEED-001: Blog Card
Each blog is represented as a card containing:
- Thumbnail/cover image
- Title
- Short excerpt (first ~100 characters of content)
- Author avatar + name
- Estimated read time
- Tags (first 2-3 shown)
- Like count
- Dislike count
- Comment count

Layout: 3 blog cards per row (grid layout)

### FR-FEED-002: Explore Tab
- Shows all public blogs from all users
- Ranking algorithm: mix of recency + engagement (likes + comments) + priority boost for writers the user follows
- Paginated (numbered pages)
- Filterable by tag or category

### FR-FEED-003: Following Tab
- Shows blogs exclusively from users the current user follows
- Ordered by most recent first
- Paginated (numbered pages)
- Only visible to signed-in users

---

## 6. Social Features

### FR-SOCIAL-001: Following
- Any user can follow any other user (one-way, no approval needed)
- Following a writer gives their blogs priority in Explore feed
- Following shows their blogs in Following tab
- User can unfollow at any time

### FR-SOCIAL-002: Friend System (Mutual)
- User A sends friend request to User B
- User B can Accept or Reject
- On Accept: both become mutual friends
- On Reject: request is removed, User A is not notified of rejection
- Friends can see each other's Friend-only blogs
- User can unfriend at any time

### FR-SOCIAL-003: Like / Dislike
- Signed-in users can like OR dislike a blog (not both simultaneously)
- Clicking the active reaction again removes it (toggle)
- Like and dislike counts displayed on blog card and blog detail page

### FR-SOCIAL-004: Comments
- Signed-in users can comment on any blog they can read
- Comments are threaded: any comment can have replies, replies can have replies (unlimited depth)
- Comment author can delete their own comment
- Moderator/Admin/Owner can delete any comment

### FR-SOCIAL-005: Sharing
- **Copy link**: copies the blog URL to clipboard
- **Share to Facebook**: opens Facebook share dialog with blog URL
- **Share to Zalo**: opens Zalo share with blog URL

### FR-SOCIAL-006: Report System
- Signed-in users can report a blog or a comment
- Reporter selects a reason (spam, inappropriate content, harassment, misinformation, other)
- Report is sent silently — reported user is NOT notified
- Notification sent to all Moderators and Admins
- Moderators/Admins can review and take action (delete content, warn user)

---

## 7. Notifications

### FR-NOTIF-001: In-App Notification Center
- Bell icon in header showing unread count
- Notification list (newest first)
- Mark as read individually or all at once

### FR-NOTIF-002: Notification Triggers
| Event | Recipient |
|-------|-----------|
| Someone likes your blog | Blog author |
| Someone dislikes your blog | Blog author |
| Someone comments on your blog | Blog author |
| Someone replies to your comment | Comment author |
| Someone follows you | Followed user |
| Someone sends you a friend request | Request recipient |
| Friend request accepted | Request sender |
| Blog/comment reported | All Moderators + Admins only |

---

## 8. Search

### FR-SEARCH-001: Universal Search
- Search bar accessible from all pages
- Searches across:
  - Blog titles
  - Blog full content (full-text)
  - Tags
  - Categories
  - Author names and usernames
- Results grouped by type (Blogs / Users / Tags)
- Paginated results (numbered pages)
- Respects privacy: friend-only and only-me blogs never appear in search for unauthorized users

---

## 9. User Profile

### FR-PROFILE-001: Profile Page
- Avatar (uploadable)
- Username
- Bio (free text, describe yourself)
- Favorite quote (free text)
- Published blogs grid (public + friend-only shown to friends, only-me hidden from all)
- Follower count
- Following count
- Friend count

### FR-PROFILE-002: Profile Editing
- User can edit their own avatar, bio, favorite quote, username

---

## 10. Admin Dashboard

### FR-ADMIN-001: Dashboard Access
- Accessible only to Admin and Owner roles
- Separate route/page not visible to regular users

### FR-ADMIN-002: Dashboard Features
- **User Management**: view all users, change roles, block/unblock accounts
- **Reports Queue**: list of all reported blogs and comments, with reason and reporter
- **Content Actions**: delete reported content directly from dashboard
- **Platform Stats**: total users, total blogs, total comments, new signups today/week/month, most liked blogs

---

## 11. Non-Functional Requirements

| NFR | Requirement |
|-----|-------------|
| Performance | Feed page loads under 2 seconds |
| Security | JWT auth, HTTPS only, OWASP Top 10 compliance |
| Scalability | PostgreSQL with proper indexing for search and feed queries |
| Image storage | Server-side storage, 5MB max per image |
| Pagination | All list views use numbered pagination |
| Mobile | React Native app with full feature parity to web |
| Email | Transactional emails: verification + password reset |
