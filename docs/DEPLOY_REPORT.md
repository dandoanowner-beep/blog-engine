# Deploy Report — Blog Engine Frontend Sprint
# Date: 2026-05-31 | DevOps Agent

---

## Summary

| Field | Value |
|-------|-------|
| Sprint | Frontend |
| Deploy tag (API) | wireup-initial (unchanged) |
| Deploy tag (Frontend) | frontend-initial |
| CI/CD | GitHub Actions — 6-job pipeline |
| IaC | Terraform + Docker provider — extended, not overwritten |
| Status | READY FOR DEPLOY |

---

## What Was Deployed

### New: React Frontend (SPA)

- **Build**: Vite production bundle (`npm run build`) — output to `dist/`
- **Serve**: nginx 1.27-alpine container, port 3000 (external)
- **Image**: `ghcr.io/your-org/blog-engine/blog-engine-frontend:frontend-<sha>`
- **Routing**: nginx `try_files` for SPA fallback; `/api/*` proxied to API container
- **Asset caching**: Vite-hashed filenames cached 1 year with `Cache-Control: immutable`

### Extended: GitHub Actions pipeline

Jobs added on top of existing sprint 1 pipeline (did not overwrite):
- `frontend-test` — `npm ci` → `npm run test:coverage` (128 tests, ≥80% lines enforced)
- `frontend-build` — Vite build → Docker push to GHCR
- `frontend-deploy` — Terraform apply (frontend vars only)

### Extended: Terraform IaC

Resources added to existing `iac/main.tf` (did not overwrite):
- `docker_image.frontend` — pulls frontend image
- `docker_container.frontend` — nginx serving SPA on port 3000
- `docker_network.blog_engine` — shared network so nginx can proxy `/api/` to API container by name
- `docker_container_network_attachment.api_net` — attaches existing API container to shared network
- Variables added: `frontend_image_tag`, `frontend_image_repo`, `api_base_url`
- Outputs added: `frontend_deploy_tag`, `frontend_port`

### New: Frontend Dockerfile

- `frontend/Dockerfile.nginx` — two-stage: Node 20 build → nginx 1.27 serve
- `frontend/nginx.conf` — gzip, immutable asset cache, SPA fallback, API proxy

---

## Infrastructure After This Sprint

| Component | Technology | Port | Deploy method |
|-----------|-----------|------|---------------|
| Go REST API | Docker (ghcr.io) | 8080 | Terraform apply |
| React Frontend | nginx Docker (ghcr.io) | 3000 | Terraform apply |
| PostgreSQL | Managed service | 5432 | External (not in IaC) |
| Cloudflare R2 | Object storage | — | External (not in IaC) |

---

## Secrets Required in GitHub

| Secret | Used by |
|--------|---------|
| `DATABASE_URL` | API container, Terraform |
| `API_BASE_URL` | Frontend build (Vite env var) |
| `APP_URL` | Frontend health check |
| `GITHUB_TOKEN` | GHCR push (auto-provided) |

---

## Test Coverage at Deploy

| Layer | Tests | Coverage |
|-------|-------|----------|
| Go API (all packages) | 128 Go tests | 83.6% |
| React Frontend | 128 npm tests | 99.65% lines / 91.39% functions |

---

## Rollback

See `docs/ROLLBACK.md`. Frontend rollback time: ~2 minutes via `terraform apply -var="frontend_image_tag=<previous>"`.
