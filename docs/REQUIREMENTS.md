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

---

## 12. Internationalization (i18n) — Vietnamese / English
# Delta Feature — 2026-06-07

### FR-I18N-001: Language Toggle
- A language toggle button is visible in the header on every page
- Supported languages: Vietnamese (VI) and English (EN)
- Default language for all visitors: **Vietnamese**
- Selected language is persisted in `localStorage` so it survives page refresh and new sessions
- Logged-in users also have their preference saved to their profile (so it syncs across devices)

### FR-I18N-002: UI Text Translation
- All static interface text is translated into both languages:
  - Navigation labels (Explore, Sign in, Get started, Logout, Admin)
  - Form labels and placeholders (Email, Password, Username, Bio, etc.)
  - Button text (Publish, Save Draft, Edit, Delete, Follow, Unfollow, Like, Dislike, Comment, Share, Report)
  - System messages and error messages (validation errors, success toasts, empty states)
  - Page headings and section titles
- Implementation: `react-i18next` with locale files `locales/vi.json` and `locales/en.json`

### FR-I18N-003: Blog Content — Bilingual Storage
- Every blog post stores two versions of its content:
  - **Vietnamese**: `title_vi`, `body_vi` — written by the author (primary)
  - **English**: `title_en`, `body_en` — auto-generated by translation service
- A `translation_status` field tracks state: `none` | `pending` | `done` | `failed`
- When a reader's language is set to EN, the blog displays `title_en` / `body_en`
- When a reader's language is set to VI, the blog displays `title_vi` / `body_vi`

### FR-I18N-004: Auto-Translation on Save/Publish
- When an author saves or publishes a blog, the system automatically translates `title_vi` + `body_vi` → `title_en` + `body_en` using Claude API (claude-sonnet-4-6)
- Translation is performed asynchronously in the background; the blog is saved immediately
- If translation succeeds, `translation_status` → `done`
- If translation fails (API error, timeout), `translation_status` → `failed`; the blog is still accessible in Vietnamese
- When an author edits and re-saves a blog, the English translation is regenerated

### FR-I18N-005: Translation Fallback
- If `translation_status` is `none` or `failed` and the reader's language is EN:
  - Display the Vietnamese content
  - Show a non-blocking notice: "English translation unavailable — showing Vietnamese"
- This applies to all blogs created before this feature was deployed

### FR-I18N-006: Blog Cards in Feed
- Blog card titles and excerpts are shown in the reader's selected language
- Falls back to Vietnamese if English translation is unavailable

### FR-I18N-007: Out of Scope (this sprint)
- Comments are not translated (dynamic, short-lived content)
- User bio and favorite quote are not translated
- No additional languages beyond VI and EN
- No UI for authors to manually edit/correct auto-translations
- Forum feature remains deferred

---

## CR-001 — Personal Blog Pivot (2026-06-10)

**Core value change** (supersedes the multi-writer model where it conflicts; see
`DESIGN_DECISIONS.md` Round 7 for the full decision record):

### FR-CR001-001: Owner-Only Authorship
- Only the **Owner** can create blogs. `POST /api/v1/blogs` requires role `owner`
  (enforced at routing level), all other roles receive 403.
- Supersedes the "User can post" row of the §2 permission matrix and FR-BLOG-001's
  multi-writer assumption. Registered users remain **readers**: comment, react,
  follow, report.

### FR-CR001-002: Navigation
- Header nav link "Explore" renamed to **"Articles"** (VI: **"Bài viết"**).
- "Write" link in the header is visible to the Owner only.

### FR-CR001-003: Single Feed
- The Feed page shows ONE article feed (the owner's published public blogs).
- The Explore/Following tab pair is removed (Following is meaningless with a single
  author). The follow system itself remains (acts as subscribe-to-owner for
  notifications).

### Unchanged by CR-001
- Guest partial read (FR-BLOG-006), privacy modes (FR-BLOG-003), comments/reactions,
  moderation roles, i18n (FR-I18N-*), registration & verification flows.

---

## CR-002 — Header Sections: Portfolio / Author / Categories / Forums (2026-06-11)

### FR-CR002-001: Portfolio
- Public page listing the owner's projects as cards (title, description, tech stack,
  repo link, demo link, thumbnail).
- `GET /api/v1/projects` — public. `POST/PATCH/DELETE /api/v1/projects[/{id}]` — owner-only
  (routing-level RBAC, same as blog creation).
- Owner manages projects from the site (no rebuild needed).
- Title required; description sanitized server-side.

### FR-CR002-002: Author Page — SUPERSEDED 2026-06-11
~~Public page rendering the owner's story as rich text (TipTap-authored HTML),
stored server-side (`GET/PUT /api/v1/about`), owner-editable inline.~~
**Superseded by owner decision (DESIGN_DECISIONS.md Round 8 amendment):** the Author
page is a **STATIC page built from the owner's design instructions** (pending —
OPEN_QUESTIONS.md OQ-005). No editing UI for anyone. The `GET/PUT /about` backend
and `site_content` table exist but are unused by the UI; their fate is decided when
the design instructions arrive.

### FR-CR002-003: Categories Browse
- Public page listing all categories with published-public article counts.
- `GET /api/v1/categories` — public.
- Clicking a category shows articles filtered by it:
  `GET /api/v1/blogs/feed?category=<slug>` (restores the filter param dropped in CR-001).
- Categories with zero published articles still appear (count 0).

### FR-CR002-004: Forums (placeholder)
- Nav link + "coming soon" page only. Real forum = backlog item S-02 (own sprint:
  BA discovery → architecture → TDD; new tables for topics/threads/replies + moderation).

### Navigation (amends FR-CR001-002)
- Header: Articles | Portfolio | Author | Categories | Forums (+ existing auth/admin links).
