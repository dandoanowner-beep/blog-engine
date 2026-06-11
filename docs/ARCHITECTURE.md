# System Architecture Document (SAD)
# Project: Blog Engine
# Version: 1.0 — 2026-05-30
# Owner: Architect Agent (sole owner of this file)

---

## 1. System Overview

A multi-tier blog platform with:
- **REST API** backend in Go
- **React** single-page application (web)
- **React Native** mobile app (iOS + Android)
- **PostgreSQL** as the primary database
- **Cloudflare R2** for image storage (S3-compatible object storage, no egress fees)

All clients communicate exclusively through the REST API. No direct database access from frontend.

---

## 2. High-Level Architecture

```
┌─────────────────────────────────────────────────────┐
│                     Clients                          │
│  ┌──────────────┐          ┌──────────────────────┐  │
│  │  React Web   │          │  React Native Mobile │  │
│  │  (Browser)   │          │  (iOS + Android)     │  │
│  └──────┬───────┘          └──────────┬───────────┘  │
└─────────┼─────────────────────────────┼──────────────┘
          │  HTTPS / REST JSON           │
          ▼                              ▼
┌─────────────────────────────────────────────────────┐
│                  Go REST API Server                  │
│                                                      │
│  ┌────────────┐  ┌────────────┐  ┌───────────────┐  │
│  │   Router   │  │ Middleware │  │   Handlers    │  │
│  │  (chi/mux) │  │ JWT Auth   │  │  (per domain) │  │
│  │            │  │ RBAC       │  │               │  │
│  │            │  │ RateLimit  │  │               │  │
│  └────────────┘  └────────────┘  └───────────────┘  │
│                                                      │
│  ┌────────────┐  ┌────────────┐  ┌───────────────┐  │
│  │  Services  │  │ Repository │  │  File Store   │  │
│  │  (biz logic│  │  (DB layer)│  │  (images)     │  │
│  └────────────┘  └────────────┘  └───────────────┘  │
└──────────────────────────┬──────────────────────────┘
                           │
          ┌────────────────┼──────────────────┐
          ▼                ▼                  ▼
  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
  │  PostgreSQL  │  │ Cloudflare   │  │  Email SMTP  │
  │  (primary DB)│  │ R2 (images)  │  │  (verify +   │
  │              │  │ S3-compatible│  │   reset)     │
  └──────────────┘  └──────────────┘  └──────────────┘
```

---

## 3. Backend Architecture (Go)

### 3.1 Package Structure

```
blog-engine/
├── cmd/
│   └── server/
│       └── main.go              ← entry point, wires everything
├── internal/
│   ├── auth/
│   │   ├── handler.go           ← register, login, verify, reset, oauth
│   │   ├── service.go           ← business logic
│   │   ├── repository.go        ← DB queries
│   │   └── jwt.go               ← token generation + validation
│   ├── blog/
│   │   ├── handler.go           ← CRUD, publish, draft
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── feed.go              ← explore + following feed queries
│   ├── social/
│   │   ├── handler.go           ← follow, friend, like, comment, share, report
│   │   ├── service.go
│   │   └── repository.go
│   ├── notification/
│   │   ├── handler.go           ← list, mark read
│   │   ├── service.go           ← create + dispatch notifications
│   │   └── repository.go
│   ├── search/
│   │   ├── handler.go
│   │   └── service.go           ← full-text search with tsvector
│   ├── user/
│   │   ├── handler.go           ← profile, edit, admin dashboard
│   │   ├── service.go
│   │   └── repository.go
│   ├── upload/
│   │   └── service.go           ← image validation + R2 upload (aws-sdk-go-v2)
│   └── middleware/
│       ├── auth.go              ← JWT validation middleware
│       ├── rbac.go              ← role-based access control
│       └── ratelimit.go         ← per-IP rate limiting
├── pkg/
│   ├── database/
│   │   └── postgres.go          ← DB connection pool
│   ├── email/
│   │   └── smtp.go              ← transactional email sender
│   └── sanitize/
│       └── html.go              ← HTML sanitizer (XSS prevention)
├── migrations/
│   └── *.sql                    ← numbered SQL migration files
└── config/
    └── config.go                ← env-based config
```

### 3.2 Request Lifecycle

```
HTTP Request
    → Router (match path + method)
    → RateLimit middleware
    → JWT Auth middleware (validates token, sets user context)
    → RBAC middleware (checks role permission)
    → Handler (parse + validate request)
    → Service (business logic, calls Repository)
    → Repository (SQL query → PostgreSQL)
    → Response (JSON)
```

### 3.3 Authentication Flow

```
Registration:
  POST /auth/register → hash password (bcrypt) → insert user →
  generate verify token → send email → return 201

Email Verification:
  GET /auth/verify?token=xxx → validate token + expiry →
  set user.verified=true → return 200

Login:
  POST /auth/login → check credentials → bcrypt compare →
  generate JWT (15min) + refresh token (7 days) → return tokens

Google OAuth:
  GET /auth/google → redirect to Google
  GET /auth/google/callback → validate Google token →
  upsert user (verified=true) → return JWT + refresh token

Refresh:
  POST /auth/refresh → validate refresh token →
  return new JWT access token

Password Reset:
  POST /auth/forgot-password → generate reset token (1hr) → send email
  POST /auth/reset-password → validate token → bcrypt new password
```

### 3.4 Feed Algorithm

Explore feed score per blog:
```
score = (like_count * 3) + (comment_count * 2) + recency_bonus + follow_boost

recency_bonus = max(0, 100 - hours_since_published * 2)
follow_boost  = 50 if current_user follows the blog author, else 0
```
Score computed on write (updated on each like/comment). Stored in `blogs.feed_score` column. Feed query orders by `feed_score DESC`.

---

## 4. Database Architecture (PostgreSQL)

See `DB_SCHEMA.md` for full table definitions.

Key design decisions:
- UUIDs as primary keys (portable, no sequential leak)
- `tsvector` column on `blogs` for full-text search (GIN index)
- Soft-delete pattern NOT used — hard delete to preserve privacy
- `feed_score` denormalized column for O(1) feed ordering
- Separate `blocks` table for mutual-blind enforcement

---

## 5. Frontend Architecture (React)

### 5.1 Tech Choices
- **React 18** with functional components + hooks
- **React Router v6** for client-side routing
- **React Query (TanStack Query)** for server state + caching
- **Zustand** for client state (auth user, notification count)
- **Tiptap** for WYSIWYG rich text editor (supports markdown, code blocks)
- **Axios** for HTTP requests
- **Tailwind CSS** for styling

### 5.2 Page Structure
```
/                    → Explore feed (public, redirects to /explore)
/explore             → Explore feed tab
/following           → Following feed tab (auth required)
/blog/:id            → Blog detail page
/blog/new            → Blog editor (auth + verified required)
/blog/:id/edit       → Blog editor (author only)
/profile/:username   → User profile page
/notifications       → Notification center (auth required)
/search?q=           → Search results page
/admin               → Admin dashboard (Admin + Owner only)
/auth/login          → Login page
/auth/register       → Registration page
/auth/verify         → Email verification landing
/auth/reset-password → Password reset page
```

### 5.3 Auth State
- JWT stored in `httpOnly` cookie (XSS-safe)
- Refresh token in `httpOnly` cookie
- User role + ID stored in Zustand after login
- All API requests attach JWT via axios interceptor

---

## 6. Mobile Architecture (React Native)

- **React Native 0.73+** (bare workflow)
- Shares API client code with web (separate package)
- **React Navigation** for screen routing
- Same Zustand store pattern for auth state
- Image upload uses React Native image picker + form-data
- Full feature parity with web (Sprint 4)

---

## 7. Security Architecture

| Threat | Mitigation |
|--------|-----------|
| XSS via blog content | Server-side HTML sanitization (bluemonday) before storage and on render |
| SQL Injection | All queries use parameterized statements (pgx library) |
| JWT theft | Short-lived access tokens (15min), httpOnly cookie storage |
| Brute force login | 5 failed attempts → 15min lockout per account |
| Image upload abuse | Type validation (JPEG/PNG/WEBP only) + 5MB size limit |
| CSRF | SameSite=Strict cookie policy + CORS whitelist |
| Unauthorized access | RBAC middleware on every protected route |
| Report spam | One report per user per content enforced at DB unique constraint |

---

## 8. Non-Functional Architecture

| Concern | Decision |
|---------|----------|
| Performance | Feed score denormalized; PostgreSQL connection pool (max 25 conns) |
| Scalability | Stateless API (JWT) — horizontal scaling ready |
| Image storage | Cloudflare R2 (S3-compatible, no egress fees) — see ADR-007 |
| Email | SMTP via configurable provider (env vars: SMTP_HOST, SMTP_PORT, SMTP_USER) |
| Migrations | Numbered SQL files in `migrations/`, applied on startup |
| Config | All config via environment variables, no hardcoded values |
| Logging | Structured JSON logs to stdout (level: info/warn/error) |

---

## 9. i18n Architecture — Bilingual VI/EN (Delta — 2026-06-07)

### 9.1 Overview

Two-layer i18n:
- **Layer 1 — UI text**: `react-i18next` on the frontend. All static strings resolved from locale JSON files at runtime. Language stored in `localStorage`.
- **Layer 2 — Blog content**: Bilingual columns on the `blogs` table (`title_en`, `body_en`). Auto-populated via Claude API at write time. Displayed based on reader's language preference.

### 9.2 Translation Service (Backend)

New package: `internal/translation/`

```
internal/translation/
├── service.go      ← Translator interface + ClaudeTranslator implementation
└── service_test.go ← unit tests (mock HTTP transport)
```

**Interface**:
```go
type Translator interface {
    Translate(ctx context.Context, titleVI, bodyVI string) (titleEN, bodyEN string, err error)
}
```

**Flow — blog create**:
```
POST /blogs
  → blog saved to DB (translation_status = 'pending')
  → HTTP 201 returned to client immediately
  → goroutine launched: translateAndUpdate(blogID, titleVI, bodyVI)
      → ClaudeTranslator.Translate() → POST api.anthropic.com/v1/messages
      → on success: UPDATE blogs SET title_en=?, body_en=?, translation_status='done'
      → on failure: UPDATE blogs SET translation_status='failed'; log error
```

**Flow — blog edit** (only if title or body changed):
```
PATCH /blogs/:id
  → detect if req.Title != existing.Title || req.Content != existing.Content
  → save update to DB (translation_status = 'pending')
  → HTTP 200 returned immediately
  → goroutine: re-translate and update title_en, body_en
```

**Claude API prompt** (claude-sonnet-4-6):
```
System: You are a professional Vietnamese-to-English translator.
User: Translate this blog post from Vietnamese to English.
      Return ONLY a JSON object with exactly two fields: "title" and "body".
      Do not include any other text or explanation.

      Title: <title_vi>
      Body: <body_vi>
```

**Config** (new env vars):
```
ANTHROPIC_API_KEY=sk-ant-...
TRANSLATION_MODEL=claude-sonnet-4-6
```

### 9.3 Frontend i18n Architecture

**Library**: `react-i18next` + `i18next`

**Locale files**:
```
src/
  i18n.ts              ← i18next init: language detection from localStorage, fallback 'vi'
  locales/
    vi.json            ← all UI strings in Vietnamese (primary/default)
    en.json            ← all UI strings in English
```

**Language detection chain** (i18next `languageDetector`):
1. Check `localStorage['i18nextLng']`
2. Default to `'vi'`

**LanguageToggle component** (`src/components/LanguageToggle.tsx`):
- Renders `VI | EN` button pair in header
- Calls `i18n.changeLanguage(lang)` — i18next persists to localStorage automatically

**Blog content language selection**:
```typescript
// In BlogCard and BlogDetail
const { i18n } = useTranslation()
const showEN = i18n.language === 'en' && blog.translation_status === 'done'
const displayTitle   = showEN ? blog.title_en   : blog.title
const displayContent = showEN ? blog.body_en    : blog.content
```

**Unavailable notice** (renders only when `language === 'en'` and `translation_status !== 'done'`):
```tsx
{i18n.language === 'en' && blog.translation_status !== 'done' && (
  <p className="text-xs text-amber-600">{t('blog.translationUnavailable')}</p>
)}
```

### 9.4 Updated Package Structure (delta additions only)

```
internal/
  translation/         ← NEW: Claude API translation service
    service.go
    service_test.go

src/ (frontend)
  i18n.ts              ← NEW: i18next config
  locales/
    vi.json            ← NEW: Vietnamese UI strings
    en.json            ← NEW: English UI strings
  components/
    LanguageToggle.tsx ← NEW: VI/EN toggle button
```
