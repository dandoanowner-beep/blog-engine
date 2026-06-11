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

---

## Round 6 — 2026-06-07 — i18n Feature Discovery

### Language Support
- **Decision**: Support Vietnamese (VI) and English (EN) only (no other languages this sprint)
- **Default language**: Vietnamese

### Translation Scope
- **Decision**: Translate UI text AND blog content
- **Not translated**: comments, user bio, user favorite quote (out of scope this sprint)

### Blog Content Translation Approach
- **Decision**: Option D — author writes once in Vietnamese; system auto-translates to English via Claude API at write/publish time; both versions stored in DB
- **Rationale**: Quality of AI translation (Option C) + zero per-read cost + full stored control (Option A). Author never writes twice.
- **Translation model**: `claude-sonnet-4-6` (cost-efficient, high quality for VI→EN)
- **Trigger**: on blog create AND on blog edit (if title or body changed)
- **Failure behavior**: blog still saved, translation_status=failed, readers see Vietnamese with notice

### Language Preference Persistence
- **Decision**: Store in localStorage (works for guests too); optionally sync to user profile for cross-device

### Forum Feature
- **Decision**: Deferred — do NOT build Forum in this sprint
- **Rationale**: User chose to focus on i18n first; Forum is a large separate feature to plan properly later

---

## Round 7 — 2026-06-10 — CR-001: Personal Blog Pivot (core value change)

**Human instruction:** "I want to build this blog like a personal blog, not a place where
anyone can write a blog. Change the Explore button on the header."

### Core Value Change
- **Decision**: The platform is a PERSONAL blog. Only the **Owner** writes articles.
- This **supersedes** Round 1 "Regular User: signs up, posts, comments, likes, follows" —
  regular users no longer post. They remain READERS: register, comment, react, follow.

### Scope (human chose "Full pivot")
1. **Backend**: only `owner` role can create blogs — enforced at API routing level
   (`POST /blogs` behind `RequireRole("owner")`), consistent with the BUG-006
   middleware-first architecture.
2. **Header nav**: "Explore" → **"Articles"** (VI: **"Bài viết"**) — emphasizes the
   article archive of a single author.
3. **Feed page**: Explore/Following tabs removed → single article feed. The Following
   tab made sense only in a multi-writer world.
4. **Registration stays** — readers sign up to comment/react (Round 1 guest gate and
   comment rules unchanged).
5. **Social features kept for readers** (follow = subscribe to the owner, comments,
   reactions, blocks). NOT removed this round — human explicitly chose "Full pivot"
   over "Pivot + drop social".

### Explicitly unchanged
- Guest partial-read gate (FR-BLOG-006 / BUG-006 fix)
- Moderator/Admin roles (still needed for comment moderation)
- i18n behavior
- ~~Backend endpoint path `/blogs/feed/explore` kept as-is (renaming the route is API
  churn with no product value; logged as tech debt)~~ **Resolved same day:** human chose
  to close the debt — route renamed to `GET /blogs/feed`, following-feed route removed,
  and the feed handler wired to the repository (see TECHNICAL_NOTES.md 2026-06-10 CR-001
  tech debt entry and BUGS.md BUG-007)

### Follow-ups (open, not blocking)
- `/editor` page is still reachable by URL for non-owners — the create API now rejects
  them (403), but the page could hide/redirect more gracefully later.
- Mobile sprint (S-01) scope must inherit this pivot.

---

## Round 8 — 2026-06-11 — CR-002: Four new header sections

**Human instruction:** Add to the header:
- **Portfolio** — "where I public my project in here"
- **Author** — "where I tell a story about me"
- **Categories** — "where user can find a specific blog to read"
- **Forums** — "where other user can share their though through some topic"

### Discovery decisions
1. **Forums scope**: placeholder page + nav link THIS round; the real forum
   (topics/threads/replies/moderation) is its own planned sprint — re-confirms the
   Round 6 deferral. Added to backlog as S-02.
2. **Portfolio source**: owner-managed in DB — new `projects` table, owner-only CRUD
   API, edit UI on the site itself (consistent with how blogs work).
3. **Author page content**: dedicated owner-editable rich-text document (TipTap, same
   editor stack as blogs), stored server-side — NOT the profile bio (more expressive).
4. **Categories page**: browse list of categories (with article counts) → click
   filters the article feed (`GET /blogs/feed?category=slug`). Restores the category
   filter dropped during the CR-001 feed wire-up.

### Nav order
`Articles | Portfolio | Author | Categories | Forums` (VI: Bài viết | Dự án | Tác giả |
Danh mục | Diễn đàn)

### Header styling — 2026-06-11 (owner instructions, iterated live)
- Spacing between the six header items (Bài viết → Dự án → Tác giả → Danh mục →
  Diễn đàn → VN toggle): owner iterated 26px → 20px → **final: 40px**.
- The five page links are **bold** (`font-bold`).
- Implementation: links wrapped in a flex group `gap-[40px] mr-[24px]` (nav's base
  gap-4 = 16px + 24px margin = 40px before the VN toggle) in `Layout.tsx`.

### Amendment — 2026-06-11 (owner UI review)
Process note: the Author "Edit" affordance and the "Viết bài" placement were built
without asking the owner first — owner objected; rule recorded (never guess UI, ask
with options/mockups). Owner's decisions when asked:
- **"Viết bài"**: KEEP in the header where it is.
- **Author page**: REMOVE the Edit button and inline rich-text editing entirely.
  The page becomes a **static page built from the owner's design instructions**,
  which the owner will provide later (logged in OPEN_QUESTIONS.md). Interim: minimal
  placeholder, no editing UI.
- Backend `GET/PUT /about` + `site_content` table are kept but currently unused by
  the UI — keep/remove will be decided when the design instructions arrive.

### Out of scope this round
- Real forum implementation (S-02 backlog)
- Tag-based browsing (categories only, as asked)
- Portfolio/About i18n auto-translation (owner writes in one language; can be a later delta)
