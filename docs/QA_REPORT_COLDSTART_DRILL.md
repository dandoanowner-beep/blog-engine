# QA Cold-Start Sub-Agent Drill — 2026-06-10

## Purpose

Verify the new sub-agent flow (`.claude/commands/QA.md` Step 2, see `sub-agents.md`):
a planted defect was inserted, a cold QA sub-agent was spawned with NO knowledge of the
change, and we measured whether it found the defect independently.

## Drill Setup

- **Planted defect:** `internal/social/service.go:206` — ownership check in `DeleteComment`
  inverted (`==` → `!=`). Effect: authors forbidden from deleting own comments; any
  non-author with role "user" could delete anyone's comment (broken access control / IDOR).
- **Sub-agent prompt:** file paths + commands only. No mention of the change, its location,
  or that this was a drill.
- **Defect reverted after drill** — `go test -count=1 ./internal/social/...` green. The
  codebase is in its original state.

## Result: PASS

The cold sub-agent independently:
1. Ran the real test suite — caught both failing tests (`TestDeleteComment_AuthorCanDelete`,
   `TestDeleteComment_NonAuthorForbidden`)
2. Identified the exact root cause: `service.go:206-207`, inverted `!=`
3. Correctly judged the TESTS right and the IMPLEMENTATION wrong (no confirmation bias)
4. Classified it CRITICAL, mapped it to AC-SOCIAL-004, and independently flagged it as
   broken access control on `DELETE /api/v1/comments/{id}`
5. Contrasted it with the correct pattern in `blog/service.go:218`

## Side Findings (UNVERIFIED — require triage)

Beyond the planted defect, the cold reviewer reported 10 additional security findings and
broad "tautological test / stub" claims, including (severity as reported):

1. CRITICAL — Block system alleged no-op: `social/handler.go:217-223` returns hard-coded
   success; `blockSvc` allegedly never wired in `cmd/server/main.go`
2. CRITICAL — `auth.AssertVerified` allegedly never called from any production path
   (unverified users can publish — AC-AUTH-001)
3. HIGH — Guest content gate (AC-BLOG-003) enforced client-side only; full content in JSON
   — **CONFIRMED & FIXED 2026-06-10** (BUG-006: server-side truncation in `GetForViewer` +
   `OptionalAuthenticate` middleware on public read routes; see BUGS.md)
4. HIGH — React/CreateComment pass `uuid.Nil` as author ID → notifications never reach authors
5. MEDIUM — Upload type validation trusts client Content-Type (no magic-byte check)
6. MEDIUM — No token revocation; `/auth/refresh` route allegedly absent from main.go
7. MEDIUM — CreateComment performs no privacy/readability check on target blog
8. LOW — `ValidateRefreshToken` keyfunc doesn't pin signing method
9. LOW — `ReportExists` swallows Scan errors
10. LOW — `Unfriend` handler ignores parse + service errors

Plus: feed/search/admin/user repositories alleged to be stubs with mock-only coverage,
and AC-AUTH-002 (Google OAuth) alleged unimplemented.

**Triage caveat:** these claims have NOT been verified by the master agent. Several may be
stale (the wire-up sprint addressed stub wiring) or wrong. Each must be checked against
the current code and `PIPELINE_STATUS.md` before any action. If confirmed, follow the bug
protocol (BUGS.md entry → failing test → fix).

## Conclusion

The sub-agent flow works as designed: cold start, real execution, no developer-intent
bias, correct adversarial judgment. The volume of side findings (vs. prior inline QA
passes that approved these sprints) is itself evidence of the persona-bleed problem the
sub-agent architecture exists to fix.
