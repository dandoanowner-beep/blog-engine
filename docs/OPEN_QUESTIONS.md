# Open Questions — Blog Engine
# Version: 1.0 — 2026-05-30

---

## Resolved
All discovery questions from Rounds 1–4 have been answered and documented in DESIGN_DECISIONS.md.

## Currently Open — CR-002 follow-up (2026-06-11)

### OQ-005 — Author page design instructions (OPEN)
- **Context:** Owner decided the Author (Tác giả) page is a STATIC page designed to
  their instructions — this supersedes the CR-002 owner-editable rich-text document.
  The Edit button/inline editor have been removed; the page shows a minimal
  placeholder until the design arrives.
- **Waiting on:** The owner's design instructions (layout, content, imagery, VI/EN).
  Owner will provide them when asked — request them before doing any further work on
  this page.
- **Dependent decision:** keep or remove the now-unused backend (`GET/PUT /about`,
  `internal/site` package, `site_content` table) — decide when the design arrives
  (if the design needs no server-side content, remove them).

## Currently Open — i18n Sprint (2026-06-07)
None — all i18n requirements confirmed and documented. Pipeline can proceed to Architect.

### Resolved This Round
- OQ-I18N-001: Translation scope — **RESOLVED**: UI text + blog content (title + body). Comments, bio, quote excluded.
- OQ-I18N-002: Translation approach — **RESOLVED**: Option D (auto VI→EN via Claude API at write time, stored in DB)
- OQ-I18N-003: Default language — **RESOLVED**: Vietnamese
- OQ-I18N-004: Forum — **RESOLVED**: Deferred, not this sprint

## Future Considerations (not blocking Sprint 1)
- OQ-001: Email service provider — which service for transactional emails? (SendGrid, Mailgun, AWS SES?) — Architecture decision for Architect Agent.
- ~~OQ-002: Image storage~~ — **RESOLVED 2026-05-30: Cloudflare R2 (see ADR-007)**
- OQ-003: Algorithm weights — exact scoring formula for Explore feed ranking (recency weight vs engagement weight) — can be tuned post-launch.
- OQ-004: Rate limiting — max blogs per day per user? Max comments per hour? — can default to reasonable limits and tune post-launch.
