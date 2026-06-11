# User Guideline — Blog Engine
# Sprint 1 — Updated: 2026-05-30

---

## Prerequisites

- Go 1.22+
- PostgreSQL 15+
- Cloudflare R2 bucket + credentials
- SMTP service (SendGrid, Mailgun, etc.)
- Docker (for deployment)

---

## Local Development Setup

### 1. Clone and configure

```bash
git clone <repo>
cd blog-engine
cp .env.example .env
# Fill in all values in .env (see Required Environment Variables below)
```

### 2. Create the database

```bash
psql -U postgres -c "CREATE DATABASE blog_engine"
```

### 3. Run migrations

```bash
go run ./cmd/migrate
```

### 4. Start the API server

```bash
cd code
go run ./cmd/server
# API runs on http://localhost:8080
```

### 5. Run tests

```bash
cd code
go test ./...                          # all tests
go test ./... -cover                   # with coverage
go test ./internal/auth/... -v         # single package verbose
```

---

## Required Environment Variables

```env
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/blog_engine?sslmode=disable

# JWT (generate with: openssl rand -hex 32)
JWT_SECRET=your-min-32-char-secret-here
JWT_REFRESH_SECRET=your-min-32-char-refresh-secret

# Cloudflare R2
R2_ACCOUNT_ID=your-account-id
R2_ACCESS_KEY_ID=your-access-key
R2_SECRET_ACCESS_KEY=your-secret-key
R2_BUCKET_NAME=blog-engine-images
R2_PUBLIC_URL=https://pub-xxx.r2.dev

# Email (SMTP)
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASS=your-sendgrid-api-key
SMTP_FROM=noreply@yourdomain.com

# Google OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

# App
APP_URL=http://localhost:3000
PORT=8080

# Claude API (optional — enables automatic VI→EN translation)
# Leave blank to disable translation (posts will stay translation_status="none")
ANTHROPIC_API_KEY=your-anthropic-api-key
```

---

## API Base URL

```
http://localhost:8080/api/v1
```

Key endpoints available in Sprint 1:
- `POST /auth/register` — create account
- `GET  /auth/verify?token=` — verify email
- `POST /auth/login` — get JWT tokens
- `GET  /auth/google` — Google OAuth
- `POST /auth/forgot-password` — request reset email
- `POST /blogs` — create blog (owner-only since CR-001; auth + verified required)
- `GET  /blogs/:id` — read blog (privacy enforced)
- `DELETE /blogs/:id` — delete blog
- `POST /uploads/image` — upload image to R2
- `GET  /blogs/feed` — Articles feed (public; replaces explore/following feeds per CR-001)

Full API reference: `docs/API_CONTRACT.md`

---

## Deployment

```bash
# Build Docker image
docker build -t blog-engine-api .

# Run with env file
docker run --env-file .env -p 8080:8080 blog-engine-api

# CI/CD: push to main branch triggers automatic deploy via GitHub Actions
```

See `docs/ROLLBACK.md` for rollback procedure.

---

## Sprint 2 — New Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/users/:id/follow` | Follow a user |
| DELETE | `/users/:id/follow` | Unfollow a user |
| POST | `/users/:id/friend-request` | Send friend request |
| PATCH | `/friend-requests/:id` | Accept or reject request |
| POST | `/blogs/:id/react` | Like or dislike a blog |
| DELETE | `/blogs/:id/react` | Remove reaction |
| POST | `/blogs/:id/comments` | Post a comment or reply |
| DELETE | `/comments/:id` | Delete a comment |
| POST | `/reports` | Report a blog or comment |
| GET | `/notifications` | List notifications |
| PATCH | `/notifications/:id/read` | Mark notification read |
| PATCH | `/notifications/read-all` | Mark all read |

Full reference: `docs/API_CONTRACT.md`

---

## How to Run the Server (Wire-up Sprint)

The server is now fully wired. Here's how to run it end-to-end:

```bash
# 1. Copy and fill in .env
cp code/.env.example code/.env

# 2. Create the database and run migrations
createdb blog_engine
psql $DATABASE_URL -f code/migrations/001_initial.sql

# 3. Start the server
cd code
go run ./cmd/server
# → Server running on http://localhost:8080

# 4. Health check
curl http://localhost:8080/health
# → {"status":"ok"}
```

### Quick API test with curl

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","username":"yourname","password":"password123"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"password123"}'

# Articles feed (public)
curl http://localhost:8080/api/v1/blogs/feed
```

---

---

## Quickstart — Full Stack (Docker)

The fastest way to run everything locally. Requires Docker Desktop.

```bash
# 1. Clone and enter the project
cd output/blog-engine

# 2. Start all three services (postgres + api + frontend)
docker compose up --build -d

# 3. Verify all are healthy
docker compose ps

# 4. Open the app
open http://localhost:3000
```

### First-time setup after docker compose up

**Register an account** at http://localhost:3000/auth/register, then instantly verify it (SMTP not configured locally):

```bash
docker compose exec postgres psql -U blog -d blog_engine \
  -c "UPDATE users SET verified=true;"
```

Log in and start writing. Everything works except:
- **Image uploads** — requires real Cloudflare R2 credentials in `.env.local`
- **Email sending** — requires real SMTP in `.env.local`
- **Google OAuth** — requires real Google client ID in `.env.local`

### Stop / restart

```bash
docker compose down          # stop containers (data preserved in volume)
docker compose down -v       # stop + wipe database
docker compose up -d         # restart without rebuilding
docker compose up --build -d # restart with rebuild (after code changes)
```

### Fix MTU on Windows (WSL2) if image pulls fail

```bash
wsl -d docker-desktop ip link set eth0 mtu 1450
```

---

## Sprint 3 — New Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/search?q=&page=` | Universal search (blogs, users, tags) |
| GET | `/users/:username` | Get user profile (visibility by relation) |
| PATCH | `/users/me` | Update own profile |
| GET | `/admin/users?page=&role=` | List all users (admin+) |
| PATCH | `/admin/users/:id/role` | Change user role (admin+) |
| GET | `/admin/reports?status=&page=` | Reports queue (moderator+) |
| PATCH | `/admin/reports/:id/resolve` | Resolve a report (moderator+) |
| GET | `/admin/stats` | Platform statistics (admin+) |

---

## i18n-Bilingual Sprint — Backend

### New Endpoint

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| PUT | `/users/me/language` | Required | Set language preference (`"vi"` or `"en"`) |

### Language Preference

Users can switch between Vietnamese and English:

```bash
# Set language to English
curl -X PUT http://localhost:8080/api/v1/users/me/language \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"language":"en"}'
# → {"language":"en"}

# Set back to Vietnamese
curl -X PUT http://localhost:8080/api/v1/users/me/language \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"language":"vi"}'
# Valid values: "vi" | "en" — anything else returns 400
```

### Translation Feature

When `ANTHROPIC_API_KEY` is set, blog posts are automatically translated Vietnamese → English in the background after creation or content update.

**Response fields added to `GET /blogs/:id`:**

```json
{
  "title":              "Tiêu đề bài viết",
  "content":            "Nội dung ...",
  "title_en":           "Blog post title",
  "body_en":            "Content ...",
  "translation_status": "done"
}
```

**Translation statuses:**

| Status | Meaning |
|--------|---------|
| `none` | No API key configured — translation skipped |
| `pending` | Translation goroutine in progress |
| `done` | Translation complete — `title_en` / `body_en` available |
| `failed` | Claude API error — check `ANTHROPIC_API_KEY` validity |

**Run the migration before starting the server:**

```bash
docker compose exec -T postgres psql -U blog -d blog_engine \
  -f migrations/002_i18n.sql
```

Or on first startup with the updated Docker image the migration runs automatically via the migration runner.

---

## i18n-Bilingual Sprint — Frontend

### Language Toggle

A **VI / EN** toggle appears in the navigation header for all users (guests and logged-in). Clicking a button switches the UI language immediately without page reload. The preference is stored in `localStorage` under key `blog_engine_lang` and survives browser close.

For logged-in users, language preference is also persisted in the DB via `PUT /users/me/language` so the preference follows the user across devices (requires a page with the toggle to be visited on each device).

### Blog Content Language Switching

After a blog post is translated (backend sprint — see above), readers can view it in English:

| UI state | Behavior |
|----------|----------|
| Language = VI (default) | Always shows Vietnamese title + content |
| Language = EN, `translation_status = done` | Shows English `title_en` + `body_en` |
| Language = EN, `translation_status ≠ done` | Shows VI content + amber "Translation unavailable" notice |

### Running the Updated Frontend

After pulling the i18n-frontend changes, rebuild the Docker image:

```bash
docker compose up --build -d
```

Or in dev mode:

```bash
cd frontend
npm install   # installs react-i18next + i18next
npm run dev   # starts on http://localhost:3000
```

### Adding New Locale Strings

All UI strings live in:
- `frontend/src/locales/vi.json` — Vietnamese (default)
- `frontend/src/locales/en.json` — English

Key naming convention: `namespace.component.element` (e.g. `nav.signIn`, `blog.delete`).

To add a string: add the key to **both** locale files, then use `t('your.key')` in the component via `const { t } = useTranslation()`.
