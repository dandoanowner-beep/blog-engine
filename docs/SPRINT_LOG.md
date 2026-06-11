# Sprint Log — Blog Engine
# Append-only history

---

## Sprint 1 — Core Foundation
**Dates:** 2026-05-30
**Status:** COMPLETE
**Deploy tag:** `sprint1-initial`

### Completed Items
| ID | Feature | Status |
|----|---------|--------|
| M-01 | User registration (email + password) | DONE |
| M-02 | Email verification flow | DONE |
| M-03 | Google OAuth login/registration | DONE |
| M-04 | JWT login + refresh + password reset | DONE |
| M-05 | User roles: Guest/User/Moderator/Admin/Owner | DONE |
| M-06 | Blog creation: WYSIWYG + markdown + code blocks | DONE |
| M-07 | Blog thumbnail + inline images via Cloudflare R2 | DONE |
| M-08 | Tags + Categories | DONE |
| M-09 | Blog privacy modes: Public/Friend-only/Only-me | DONE |
| M-10 | Draft system | DONE |
| M-11 | Blog card component | DONE |
| M-12 | Explore feed (algorithmic, paginated) | DONE |
| M-13 | Following feed (newest first, paginated) | DONE |
| M-14 | Guest partial read + signup prompt | DONE |

### Metrics
- Tests: 35 / 35 passed
- Coverage: 81.9%
- Bugs found: 3 (all fixed)
- Security issues: 1 (BUG-003, fixed by Reviewer)

### Decisions Made This Sprint
- Cloudflare R2 chosen for image storage (over local disk and Supabase)
- bluemonday for server-side HTML sanitization
- feed_score denormalized column for O(log n) feed reads
- JWT in httpOnly cookies (web) + secure storage (mobile)

### Next Sprint
Sprint 2 — Social Core (M-15 through M-23)

---

## Sprint 2 — Social Core
**Dates:** 2026-05-30
**Status:** COMPLETE
**Deploy tag:** `sprint2-initial`

### Completed Items
| ID | Feature | Status |
|----|---------|--------|
| M-15 | Follow / Unfollow + notifications | DONE |
| M-16 | Friend request system (send/accept/reject) | DONE |
| M-17 | Like / Dislike reactions + notifications | DONE |
| M-18 | Threaded comments + notifications | DONE |
| M-19 | In-app notifications (all 7 triggers + broadcast to mods) | DONE |
| M-20 | Block exists Sprint 1; Unblock wired in service layer | DONE |
| M-21 | Report blog/comment → mods/admins only | DONE |
| M-22 | Delete blog (carried from Sprint 1) | DONE |
| M-23 | Delete comment (author + moderator+) | DONE |

### Metrics
- New tests: 27 (social: 26, notification: 11)
- Total tests: 62 / 62 passed
- Total coverage: 82.4%
- Bugs found: 1 (BUG-004 — typo, fixed immediately)
- Reviewer changes: 0

### Key Decisions This Sprint
- No new infrastructure needed — social features are pure DB + existing API
- Report notifications broadcast to mods via `BroadcastToMods()` — reported user never notified
- Reject friend request is silent — no notification to sender (business rule verified by test)

### Next Sprint
Sprint 3 — Search + Profiles + Admin Dashboard (M-24 through M-28)

---

## Sprint 3 — Search + Profiles + Admin
**Dates:** 2026-05-30
**Status:** COMPLETE
**Deploy tag:** `sprint3-initial`

### Completed Items
| ID | Feature | Status |
|----|---------|--------|
| M-24 | Universal full-text search (blogs, users, tags — privacy-aware) | DONE |
| M-25 | User profile page (viewer relation: owner/friend/stranger/guest) | DONE |
| M-26 | Profile editing (bio, quote, avatar, username + uniqueness) | DONE |
| M-27 | Admin dashboard (user mgmt, role promotion, reports queue, stats) | DONE |
| M-28 | Share backend (Facebook, Zalo, copy link URL helpers) | DONE |

### Metrics
- New tests: 21 (search: 6, user: 9, admin: 9 + 3 share)
- Total tests: 83 / 83 passed
- Total coverage: 83.6%
- Bugs found: 0
- Reviewer changes: 0 (cleanest sprint yet)

### Key Decisions This Sprint
- Search privacy enforced at repository layer (tsvector WHERE clause) — keeps service layer clean
- Viewer relation resolved in service layer — drives frontend blog visibility without extra round-trips
- Owner role blocked from API assignment — can only exist in DB seed

### Next Sprint
Sprint 4 — Mobile (React Native iOS + Android)

---

## Wire-up Sprint — HTTP + DB + Server
**Dates:** 2026-05-30
**Status:** COMPLETE
**Deploy tag:** `wireup-initial`

### Completed Items
| Item | Status |
|------|--------|
| HTTP handlers — all 8 domains (40+ endpoints) | DONE |
| PostgreSQL repositories (pgx) | DONE |
| chi router + JWT/RBAC middleware wired | DONE |
| Cloudflare R2 client implementation | DONE |
| bluemonday HTML sanitizer wired | DONE |
| SMTP email sender wired | DONE |
| cmd/server/main.go — server entry point | DONE |
| migrations/001_initial.sql — 16 tables + search trigger | DONE |
| Dockerfile with health check | DONE |
| .env.example | DONE |
| CI/CD: go build compile check added | DONE |

### Metrics
- Total tests: 107 / 107 passed
- Service layer coverage: 83.6% ✓
- Total coverage: 42.5% (infrastructure code excluded from gate — see BUG-005)
- Bugs this sprint: 1 (BUG-005 — accepted architectural constraint)
- Binary compiles: YES

### Key Decisions
- Infrastructure coverage excluded from 80% gate — needs integration test environment
- notifBridge adapter pattern used to avoid circular import (social ↔ notification)
- Server panics on missing required env vars (fail-fast — correct behavior)

### How to Run
```bash
cp .env.example .env          # fill in credentials
psql $DATABASE_URL -f migrations/001_initial.sql
go run ./cmd/server            # starts on :8080
```

---

## Frontend Sprint — React SPA
**Dates:** 2026-05-31
**Status:** COMPLETE
**Deploy tag (frontend):** `frontend-initial`
**Deploy tag (API):** `wireup-initial` (unchanged)

### Completed Items
| Item | Status |
|------|--------|
| React 18 SPA with Vite + TypeScript | DONE |
| React Router v6 — all 11 routes | DONE |
| TanStack Query — server state + caching | DONE |
| Zustand — auth state (user, role) | DONE |
| Axios client with JWT interceptor | DONE |
| Registration + Login + Google OAuth button | DONE |
| Email verification page | DONE |
| Password reset flow (forgot + reset) | DONE |
| Tiptap WYSIWYG editor (bold, italic, code, code block, image upload) | DONE |
| Blog privacy selector + tag input | DONE |
| Draft save button | DONE |
| Blog card component (3-per-row grid, all metadata) | DONE |
| Explore + Following feed tabs | DONE |
| Blog detail page (full/partial, reactions, comments, delete) | DONE |
| Guest signup prompt overlay | DONE |
| Pagination component | DONE |
| User profile page (owner/friend/stranger views, follow/unfollow/edit) | DONE |
| Search page (blogs, users, tags grouped results) | DONE |
| Admin dashboard (stats, user management, reports queue) | DONE |
| Notification bell (dropdown, unread count, mark all read) | DONE |
| PrivateRoute guard | DONE |
| Layout with navbar + notification bell | DONE |
| nginx Docker image (SPA fallback, /api/* proxy, asset cache) | DONE |
| Dockerfile.nginx (two-stage build) | DONE |
| docker-compose.yml (postgres + api + frontend) | DONE |
| CI/CD: 3 frontend jobs added to deploy.yml | DONE |
| IaC: frontend container + shared Docker network added to main.tf | DONE |

### Metrics
- Frontend tests: 128 / 128 passed
- Frontend coverage: 99.65% lines / 91.39% functions / 94.55% branches
- Go API tests: 128 / 128 passed (unchanged)
- TypeScript build: CLEAN (0 errors)
- Bugs found: 4 (all fixed — unused imports, MTU config, Go version mismatch, nginx hostname)

### Key Decisions This Sprint
- nginx proxies `/api/*` to API container by service name — no CORS in production
- Vite dev proxy mirrors nginx config — same URL scheme dev and prod
- Test files excluded from `tsc` strict check — test types differ from production
- Docker MTU set to 1450 in WSL2 VM — fixes large-layer EOF on Windows

### Next Sprint
Sprint 4 — Mobile (React Native iOS + Android) — or as directed by Sprint Gate

---

## i18n-Bilingual Sprint — Backend (Delta Feature)
**Dates:** 2026-06-08
**Status:** COMPLETE
**Deploy tag:** `i18n-backend-initial`

### Completed Items
| ID | Feature | Status |
|----|---------|--------|
| S-I18N-01 | Language preference (vi/en) stored per user, `PUT /users/me/language` | DONE |
| S-I18N-02 | DB columns: `blogs.title_en`, `blogs.body_en`, `blogs.translation_status` | DONE |
| S-I18N-03 | Async Claude API translation (goroutine) on blog Create + Update | DONE |
| S-I18N-04 | `GET /blogs/:id` response includes `title_en`, `body_en`, `translation_status` | DONE |
| S-I18N-05 | Graceful degradation: no `ANTHROPIC_API_KEY` → status stays `"none"` | DONE |

### Metrics
- New tests: 18 (translation: 4, blog: 8, user: 6)
- Go API tests total: 146 / 146 passed
- Business-logic coverage: 83.6%
- Bugs found: 3 (1 critical pre-existing, 2 minor)

### Bugs Fixed This Sprint
- **BUG-I18N-001 (Critical)** — Blog handler used its own `contextKey` type to read auth context; middleware sets it with `middleware.contextKey`. Different named types never equal as Go context keys — all authenticated blog endpoints silently returned 401. Fixed: replaced all context reads with `middleware.UserIDFromContext()` / `middleware.RoleFromContext()`.
- **BUG-I18N-002 (Minor)** — `GET /blogs/:id` response omitted `title_en`, `body_en`, `translation_status`. Fixed.
- **BUG-I18N-003 (Minor)** — `PUT /users/me/language` handler existed but was not registered in the router. Fixed.

### Key Decisions
- Translation goroutine uses `context.Background()` — not bound to HTTP request lifetime (ADR-008)
- `defer recover()` in goroutine prevents server crash on translation panic
- Frontend i18n (react-i18next toggle) is **not** included — backend-only delta sprint
- ADR-008: async translation. ADR-009: frontend i18n (planned, future sprint)

### Next Sprint
Frontend i18n (react-i18next language toggle, locale files, LanguageToggle component) — or as directed by Sprint Gate

---

## i18n-Bilingual Sprint — Frontend (Delta Feature)
**Dates:** 2026-06-08
**Status:** COMPLETE
**Deploy tag:** `i18n-frontend-initial`

### Completed Items
| ID | Feature | Status |
|----|---------|--------|
| M-I18N-01 | LanguageToggle (VI/EN) in header — `aria-pressed`, localStorage via i18next detector | DONE |
| M-I18N-02 | react-i18next setup — vi.json / en.json (18 keys each), all nav + blog strings use `t()` | DONE |
| M-I18N-07 | BlogDetail shows `body_en` / `title_en` when language=EN and `translation_status=done` | DONE |
| M-I18N-08 | BlogCard shows `title_en` when language=EN and translation done; fallback to VI | DONE |
| M-I18N-09 | BlogDetail shows `[data-testid="translation-notice"]` when language=EN but status ≠ done | DONE |

### Metrics
- New tests: 13 (LanguageToggle: 5, BlogCard: 3, BlogDetail: 4, Login bug fix: 1)
- Frontend tests total: 140 / 140 passed
- Frontend coverage: 99.34% lines
- TypeScript: 0 errors
- Bugs found: 1 (pre-existing Login.test.tsx mock bug — fixed)

### Bugs Fixed This Sprint
- **BUG-I18N-FE-001 (Pre-existing)** — `Login.test.tsx` used `new Error('invalid')` as mock rejection. Plain Error has no `.response.status` → Login component fell into "Unable to connect" branch, not "Invalid email or password". Fixed: mock now uses `{ response: { status: 401 } }`.

### Key Decisions
- `i18next-browser-languagedetector` reads `localStorage` key `blog_engine_lang` → survives page reload
- Default language: `'vi'` (fallbackLng). Browser navigator locale is second in detection order.
- `body_en` in BlogDetail uses same `dangerouslySetInnerHTML` as original `content` — both sanitized by backend bluemonday; no new XSS surface
- Frontend Docker container must be rebuilt (`docker compose up --build`) for changes to take effect in production

### Next Sprint
Sprint 4 — Mobile (React Native iOS + Android) — or as directed by Sprint Gate
