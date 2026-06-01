# Open Questions — Blog Engine
# Version: 1.0 — 2026-05-30

---

## Resolved
All discovery questions from Rounds 1–4 have been answered and documented in DESIGN_DECISIONS.md.

## Currently Open
None — all blockers resolved. Pipeline can proceed to PM.

## Future Considerations (not blocking Sprint 1)
- OQ-001: Email service provider — which service for transactional emails? (SendGrid, Mailgun, AWS SES?) — Architecture decision for Architect Agent.
- ~~OQ-002: Image storage~~ — **RESOLVED 2026-05-30: Cloudflare R2 (see ADR-007)**
- OQ-003: Algorithm weights — exact scoring formula for Explore feed ranking (recency weight vs engagement weight) — can be tuned post-launch.
- OQ-004: Rate limiting — max blogs per day per user? Max comments per hour? — can default to reasonable limits and tune post-launch.
