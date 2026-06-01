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
- `POST /blogs` — create blog (auth + verified required)
- `GET  /blogs/:id` — read blog (privacy enforced)
- `DELETE /blogs/:id` — delete blog
- `POST /uploads/image` — upload image to R2
- `GET  /blogs/feed/explore` — Explore feed
- `GET  /blogs/feed/following` — Following feed (auth required)

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

# Explore feed (public)
curl http://localhost:8080/api/v1/blogs/feed/explore
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
