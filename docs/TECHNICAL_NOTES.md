# Technical Notes — Blog Engine

Instructions and business logic explained by the human mid-session. Each entry is binding.

---

## 2026-06-10 — Content access architecture: server-side enforcement (BUG-006)

**Human instruction (verbatim intent):** Refactor the content access architecture. The guest
content gate must NOT be handled client-side. Implement a strict server-side middleware that
validates the session token and blocks unauthorized access at the API routing level, before
any controller logic executes. Apply it to all protected feed endpoints.

**Decisions made implementing this:**

1. **Middleware location:** `internal/middleware/auth.go` (alongside the existing
   `Authenticate` and RBAC middleware), not a new `auth/middleware.go` or
   `blog/middleware.go` — consistent with the established package layout (ADR pattern:
   cross-cutting HTTP concerns live in `internal/middleware`).
2. **Two enforcement layers:**
   - `Authenticate` (existing, strict) — applied to `/blogs/feed/following` and all other
     authenticated routes. Blocks with 401 before controller logic. Verified already in place.
   - `OptionalAuthenticate` (new) — for guest-allowed read routes (`/blogs/feed/explore`,
     `/blogs/{id}`). No token → request proceeds as guest. Token present but
     invalid/expired → **401** (strict mode: a presented credential must be valid; it is
     never silently downgraded to guest). Valid token → claims injected into context.
3. **Guest partial read (FR-BLOG-006) enforced in the service layer:** `GetForViewer`
   returns a *copy* of the blog with `Content` truncated to the first ~30% for guest
   viewers. The full content never leaves the server for a guest.
4. **Truncation operates on plain text** (HTML stripped first) so truncation can never emit
   broken/unbalanced HTML. Guest preview is plain text by design; the frontend overlay
   (`GuestPrompt.tsx`) remains as UX on top.
5. **`body_en` is truncated identically** — otherwise the full English translation would
   leak the gated content.

**Open follow-ups** (logged, not blocking): feed handlers (`feed_handler.go`) are still
repository-disconnected stubs; remaining unverified cold-start drill findings (see
`QA_REPORT_COLDSTART_DRILL.md`) still need triage.

---

## 2026-06-10 — CR-001: Personal blog pivot

See `DESIGN_DECISIONS.md` Round 7 for the full decision record. Implementation notes:

- `POST /blogs` moved into an owner-only routing group (`Authenticate` +
  `RequireRole("owner")`) in `cmd/server/main.go`. `PATCH`/`DELETE /blogs/{id}` stay in
  the general authenticated group — service-level author/moderator checks already cover
  them (non-owners have no blogs to edit; moderators keep delete for moderation).
- Frontend: `nav.explore` i18n key replaced by `nav.articles` ("Articles" / "Bài viết");
  Write link rendered only for `role === 'owner'`; `Feed.tsx` reduced to a single
  article feed (tabs removed).

---

## 2026-06-10 — CR-001 tech debt closed + feed wire-up (BUG-007)

Human instruction: "handle the tech debt first, then rebuild docker." Human confirmed
scope = route rename AND feed wire-up (homepage must actually list articles after rebuild).

- **Route rename**: `GET /blogs/feed/explore` → **`GET /blogs/feed`** (public,
  `OptionalAuthenticate`). `GET /blogs/feed/following` **removed** — meaningless with a
  single author (the follow system itself stays: subscribe → notifications).
- **Feed wire-up (was drill side finding "feed stubs")**: `Handler.ArticlesFeed` →
  `Service.ArticlesFeed` (clamps page, `ArticlesPerPage = 9` to match the frontend grid)
  → `PostgresRepository.GetArticlesFeed` (JOIN users for author username/avatar,
  `json_agg` for tags, published+public only, feed_score DESC). Blog model gained
  denormalized `AuthorUsername`/`AuthorAvatarURL` for feed cards.
- **Deleted**: orphaned `feed_repository.go` (`FeedPostgresRepository` was never
  constructed in main.go — dead code), stub `ExploreFeed`/`FollowingFeed` handlers.
- Frontend: `blogsApi.getArticlesFeed(page)` hits `/blogs/feed`; `getFollowingFeed`
  removed. Tag/category filter params dropped from the API client (backend never
  implemented them; re-add with a real filter feature).
- Feed JSON always emits `"blogs": []` (never null) — frontend `.map` would crash.

---

## 2026-06-11 — CR-002: Portfolio / Author / Categories / Forums

Decisions in DESIGN_DECISIONS.md Round 8; contract in API_CONTRACT.md CR-002 section.
Implementation notes:

- **New packages**: `internal/portfolio` (projects CRUD) and `internal/site`
  (owner-authored site documents, key-based — `about` for now). Both follow the
  blog package layout (model/service/repository/handler, sanitizer injected).
- **Migration 003**: `projects` + `site_content` tables. ⚠ initdb only runs on an
  empty volume — applied manually to the running postgres via psql at deploy.
- **Categories were dead code before this** (drill adjacent): `category_ids` plumbed
  end-to-end but `UpsertCategories` was a no-op, never called, Editor sent `[]`, no
  blog↔category rows ever written. Replaced with **name-based categories like tags**:
  `category_names` in blog create/update → upsert by slug → `SetBlogCategories`
  replaces associations. Editor gained a "Categories (comma-separated)" input.
- Feed category filter: `EXISTS` subquery on `blog_categories`/`categories.slug`;
  empty string matches all (single query, no branching).
- Owner-only writes (`POST/PATCH/DELETE /projects`, `PUT /about`) share the CR-001
  owner RBAC routing group. Public reads are plain routes.
- Forum: `/forum` placeholder page only. Real forum logged as backlog **S-02**.
