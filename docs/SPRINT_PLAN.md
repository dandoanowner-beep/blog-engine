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

---

## Sprint i18n — Bilingual Support (Vietnamese / English)
# Delta Feature Sprint — 2026-06-07
**Goal:** Any visitor can toggle between Vietnamese and English. All UI text renders in the selected language. Blog content (title + body) is auto-translated VI→EN by Claude API at publish time and stored; readers see content in their preferred language with a graceful fallback to Vietnamese when translation is unavailable.

### MoSCoW Backlog

| Priority | ID | Feature | Effort |
|----------|----|---------|--------|
| **Must** | M-I18N-01 | Language toggle (VI/EN) in header — localStorage persistence, default VI | S |
| **Must** | M-I18N-02 | UI text translation via react-i18next — all pages, components, error messages | L |
| **Must** | M-I18N-03 | DB migration: `title_en`, `body_en`, `translation_status` columns on blogs | S |
| **Must** | M-I18N-04 | Translation service (Go): Claude API `claude-sonnet-4-6` VI→EN | M |
| **Must** | M-I18N-05 | Auto-translate on blog create (async goroutine, non-blocking) | S |
| **Must** | M-I18N-06 | Auto-re-translate on blog edit (if title or body changed) | S |
| **Must** | M-I18N-07 | Blog detail: show EN content when language=EN and translation=done | S |
| **Must** | M-I18N-08 | Blog cards: show EN title/excerpt when language=EN and translation available | S |
| **Must** | M-I18N-09 | Fallback to VI content + "Translation unavailable" notice | S |
| **Should** | S-I18N-01 | Language preference synced to user profile in DB (cross-device for logged-in users) | M |
| **Could** | C-I18N-01 | Translation status badge on author's draft/blog management view | S |
| **Won't** | W-I18N-01 | Auto-detect browser language | — |
| **Won't** | W-I18N-02 | Additional languages beyond VI and EN | — |
| **Won't** | W-I18N-03 | Comment, bio, or quote translation | — |
| **Won't** | W-I18N-04 | Manual correction of auto-translations | — |

---

### Backend Tasks (Go)
| Task | Stories | AC | Effort |
|------|---------|-----|--------|
| DB migration 002: add `title_en`, `body_en TEXT`, `translation_status VARCHAR(20) DEFAULT 'none'` to blogs | US-021 | AC-I18N-003 | S |
| Update Blog model struct + repository (pgx read/write new fields) | US-020 | AC-I18N-003 | S |
| Translation service: `internal/translation/service.go` — Claude API client, translate(titleVI, bodyVI) → (titleEN, bodyEN) | US-021 | AC-I18N-004 | M |
| Wire translation into blog create: fire async goroutine after blog saved | US-021 | AC-I18N-004 | S |
| Wire translation into blog update: re-translate if title_vi or body_vi changed | US-021 | AC-I18N-004 | S |
| Update blog API response: include `title_en`, `body_en`, `translation_status` in JSON | US-020 | AC-I18N-005 | S |
| **TDD — Red**: write failing tests for translation service (mock HTTP), blog create trigger, blog update re-trigger | US-021 | AC-I18N-004 | M |
| **TDD — Green**: implement until all tests pass | — | — | M |
| **TDD — Refactor**: clean up, ensure coverage ≥ 80% on new code | — | — | S |

### Frontend Tasks (React/TypeScript)
| Task | Stories | AC | Effort |
|------|---------|-----|--------|
| Install react-i18next, i18next; create `src/i18n.ts` config | US-019 | AC-I18N-002 | S |
| Create `src/locales/vi.json` — all static UI strings in Vietnamese | US-019 | AC-I18N-002 | M |
| Create `src/locales/en.json` — all static UI strings in English | US-019 | AC-I18N-002 | M |
| LanguageToggle component — shows VI/EN, persists to localStorage | US-019 | AC-I18N-001 | S |
| Add LanguageToggle to Layout.tsx header | US-019 | AC-I18N-001 | S |
| Replace all hardcoded strings with `t('key')` across all pages/components | US-019 | AC-I18N-002 | L |
| Update API types: add `title_en`, `body_en`, `translation_status` to Blog type | US-020 | AC-I18N-003 | S |
| Update BlogCard: show EN title/excerpt when language=EN and translation=done, fallback to VI | US-020 | AC-I18N-005 | S |
| Update BlogDetail: show EN title+body when language=EN, fallback + notice when unavailable | US-020 | AC-I18N-005, AC-I18N-006 | S |
| **TDD — Red**: failing tests for LanguageToggle, BlogCard language switching, BlogDetail language switching | US-019, US-020 | AC-I18N-001, AC-I18N-005 | M |
| **TDD — Green**: implement until all tests pass | — | — | M |
| **TDD — Refactor**: clean up, ensure coverage ≥ 80% on new code | — | — | S |

### Sprint Success Criteria
- All Must-have items (M-I18N-01 through M-I18N-09) implemented and tested
- Language toggle works on all pages
- At least one blog post demonstrably shows EN content after publish
- Frontend + backend test coverage ≥ 80% on new code
- TypeScript build clean (0 errors)
- Go binary compiles
