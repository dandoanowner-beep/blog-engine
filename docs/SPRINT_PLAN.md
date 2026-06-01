# Sprint Plan — Blog Engine
# Version: 1.0 — 2026-05-30

---

## Sprint 1 — Core Foundation
**Goal:** A user can register, verify email, log in, create and publish a blog with images/tags/categories, and browse the Explore and Following feeds. Guests can read a partial preview.

### Backend Tasks (Go)
| Task | Stories | AC |
|------|---------|-----|
| User model + PostgreSQL schema (users, roles) | US-001, US-002 | AC-AUTH-001, AC-AUTH-002 |
| Email registration endpoint + password hashing | US-001 | AC-AUTH-001 |
| Email verification endpoint + token | US-003 | AC-AUTH-003 |
| Google OAuth integration | US-002 | AC-AUTH-002 |
| JWT login + refresh token endpoint | US-004 | AC-AUTH-004 |
| Password reset endpoint + email | US-004 | AC-AUTH-005 |
| Role-based middleware (RBAC) | US-018 | AC-AUTH-001 |
| Blog model + schema (blogs, tags, categories, images) | US-006 | AC-BLOG-001 |
| Blog CRUD endpoints (create, read, update) | US-006, US-007 | AC-BLOG-001, AC-BLOG-002 |
| Image upload endpoint (5MB validation, storage) | US-006 | AC-BLOG-001 |
| Privacy filter middleware | US-008 | AC-BLOG-003 |
| Explore feed endpoint (algorithm + pagination) | US-010 | AC-FEED-002 |
| Following feed endpoint (pagination) | US-011 | AC-FEED-003 |
| Guest partial-read logic (30% content slice) | US-008 | AC-BLOG-003 |

### Frontend Tasks (React)
| Task | Stories |
|------|---------|
| Registration + login forms (email + Google button) | US-001, US-002 |
| Email verification page | US-003 |
| Password reset flow pages | US-004 |
| Blog editor (WYSIWYG + markdown + code blocks + image upload) | US-006 |
| Privacy selector + tag/category picker | US-006 |
| Draft save button | US-007 |
| Blog card component (all metadata, 3-per-row grid) | AC-FEED-001 |
| Explore feed page (tabs: Explore / Following) | US-010, US-011 |
| Blog detail page (full content for signed-in, partial for guests) | US-008 |
| Guest signup prompt overlay | US-008 |
| Pagination component (numbered pages) | AC-FEED-002 |

### TDD Requirement
- Red → Green → Refactor for every endpoint
- Min 80% coverage before Gate 4

### Sprint 1 Exit Criteria
- User can register, verify, and log in
- User can create, publish, and view a blog
- Explore and Following feeds work and paginate
- Guests see partial content with signup prompt
- All Must-have items M-01 through M-14 complete

---

## Sprint 2 — Social Core
**Goal:** Users can follow, friend, react, comment, get notified, block others, and report content.

### Key Items: M-15 through M-23
- Follow/unfollow + notification
- Friend request flow + notification
- Like / Dislike toggle + notification
- Threaded comment create/delete + notification
- In-app notification center (bell icon, all 7 triggers)
- Block/unblock (mutual-blind feed filtering)
- Report blog/comment (silent, notify Moderators + Admins)
- Blog + comment delete (author + Moderator/Admin)

---

## Sprint 3 — Discovery + Admin
**Goal:** Full-text search, complete user profiles, admin dashboard, sharing.

### Key Items: M-24 through M-28
- PostgreSQL full-text search across blogs, users, tags, categories
- Search results grouped and paginated
- Privacy-aware search filtering
- User profile page (all fields, blog grid, counts)
- Profile editing
- Admin dashboard (user management, reports queue, stats)
- Share buttons (copy link, Facebook, Zalo)

---

## Sprint 4 — Mobile
**Goal:** React Native app with full feature parity to web.

### Key Items: S-01
- React Native project setup (Expo or bare)
- Shared API layer with web
- All screens from Sprint 1–3 implemented in React Native
- iOS + Android tested
