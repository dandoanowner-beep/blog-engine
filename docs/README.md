# Blog Engine

A full-featured multi-user blog platform.

## Tech Stack
- **Backend:** Go — REST API
- **Frontend:** React (web)
- **Mobile:** React Native (iOS + Android)
- **Database:** PostgreSQL

## Quick Start (Sprint 1)

```bash
# Backend
cp .env.example .env          # fill in DB + SMTP + Google OAuth credentials
go run ./cmd/server           # starts on :8080

# Frontend
cd frontend && npm install && npm start   # starts on :3000

# Database
psql -U postgres -c "CREATE DATABASE blog_engine"
go run ./cmd/migrate           # runs all migrations
```

## Key Features
- Rich blog editor (WYSIWYG + markdown + code blocks)
- 5 user roles with granular permissions
- Public / Friend-only / Only-me privacy per blog
- Algorithmic Explore feed + Following feed
- Mutual friends system + one-way following
- Threaded comments, like/dislike reactions
- In-app notifications
- Full-text search (blogs, users, tags)
- Admin dashboard with user management + reports queue
- React Native mobile app

## Documentation
- `docs/REQUIREMENTS.md` — Full requirements
- `docs/ARCHITECTURE.md` — System design
- `docs/API_CONTRACT.md` — All API endpoints
- `docs/DB_SCHEMA.md` — Database tables
- `docs/adrs/` — Architecture Decision Records
- `docs/SPRINT_PLAN.md` — Sprint breakdown
