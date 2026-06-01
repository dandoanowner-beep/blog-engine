# Design Decisions — Blog Engine

## Round 1 — 2026-05-30

### Roles
- **Owner** (you): full control over the platform
- **Admin**: can manage users, delete any post/comment
- **Moderator**: can moderate content (TBD scope — ask Round 2)
- **Regular User**: signs up, posts, comments, likes, follows
- **Guest**: can read public posts without signing up

### Content Types
- Plain text
- Images
- Markdown formatting
- Code blocks
- WYSIWYG editor (rich text, formatting toolbar like Word / Google Docs)

### Social Features
- Comments on posts
- Likes / reactions on posts
- Following other users
- Sharing posts
- Notifications (likes, comments, follows)

### Authentication
- Email + password
- Google OAuth

### Post Privacy (3 modes — writer chooses per blog)
- **only-me**: only the writer can see it (draft-like but published)
- **friend-only**: only friends of the writer can see it
- **public**: everyone including guests can see it

### Roles & Permissions
- Moderator: CAN delete posts
- Admin: can promote users to Moderator
- Owner: full platform control (above Admin)

### Content & Editor
- Images: upload from device only, max 5MB per image
- Drafts: yes — users can save before publishing
- Tags + Categories: mandatory on every blog, users can browse/filter by tag or category

### Friends vs Following (TWO separate systems)
- **Following** (one-way): user follows a writer → writer's blogs get priority in Explore feed. No mutual requirement.
- **Friends** (mutual): both users must accept a friend request. "Friend-only" blogs are visible ONLY to mutual friends.
- These are independent — you can follow someone without being their friend.

### Feed & UI
- Unit is a "blog card" (small square card), 3 cards inline per row
- Blog card contains: cover image, title, short excerpt, author name + avatar, read time estimate, tags, like count, dislike count, comment count
- Two feed tabs:
  - **Explore**: global feed of all public blogs. Followed writers get priority ranking.
  - **Following**: blogs exclusively from people the user follows

### Guest Experience
- Guests can browse and open public blog cards
- Guests can read only the top portion of a blog
- Lower content is hidden behind a signup/login prompt
- Guests cannot comment

### Comments & Reports
- Comments: signed-in users only (guests cannot comment)
- Threaded comments (reply to a reply)
- Users can report a blog OR a comment to Moderators

### Account Rules
- Users CANNOT delete their account
- Users CAN delete their own blogs
- On blog deletion: blog and all its comments/reactions are removed

### User Profile Page
- Avatar
- Bio (about themselves)
- Favorite quote
- Their published blogs
- Follower count / Following count
- Friend count

### Tech Stack (decided)
- Backend: Go
- Frontend Web: React
- Mobile: React Native (iOS + Android, one codebase)
- Database: PostgreSQL

## Architecture Change — 2026-05-30 (during Gate 2 review)
- **Image storage**: Changed from local filesystem → Cloudflare R2 (S3-compatible object storage)
- **Why**: Owner prefers not to use local disk for images
- **Impact**: Upload service uses R2 SDK instead of local FS. Images served via R2 public URL / Cloudflare CDN.

## Round 4 — 2026-05-30

### Block System
- Users can block each other
- Blocking is mutual-blind: neither sees the other's blogs in any feed
- Blocked user cannot see blocker's profile or content

### Friend Request
- Flow: Send → Accept or Reject (no ignore/pending limbo)

### Explore Feed Ranking
- Algorithm: mix of recency + engagement (likes + comments) + followed writers priority

### Categories
- Both: predefined base categories (created by Owner/Admin) + users can create their own

### Images in Blog Body
- One thumbnail/cover image (displays on blog card)
- Multiple images can be inserted anywhere in the blog body

### Notifications (full trigger list)
- Someone likes your blog
- Someone dislikes your blog
- Someone comments on your blog
- Someone replies to your comment
- Someone follows you
- Someone sends you a friend request
- Friend request accepted
- Report received → notify Moderator + Admin only (NOT the reported user)

### Admin Dashboard
- Separate private dashboard for Owner + Admin
- Features: manage users, review reported content, platform statistics

### Signup & Security
- Email verification required before user can publish a blog
- Password reset via email link (standard flow)

### Search Scope
- Blog titles
- Blog full content
- Tags and categories
- Author names / usernames

### Mobile App
- Full feature parity with web (React Native)

### Pagination
- Numbered pages (not infinite scroll)

### Social Features
- Comments: threaded (reply to a reply)
- Reactions: Like / Dislike only (2 states)
- Share: copy link to post + share to Facebook + share to Zalo
- Notifications: in-app only (bell icon)

### Search
- Critical feature — users can search posts, users, tags

### Platform
- Web browser (primary)
- Mobile app: iOS + Android

### Authentication
- Email + password
- Google OAuth
