# Risk Register — Blog Engine
# Version: 1.0 — 2026-05-30

| ID | Risk | Probability | Impact | Mitigation |
|----|------|-------------|--------|------------|
| R-01 | WYSIWYG editor XSS vulnerability (user-submitted HTML rendered in browser) | High | Critical | Sanitize all HTML output server-side (use bluemonday or similar Go library). Never trust client HTML. |
| R-02 | Image storage fills disk / high storage cost | Medium | High | Enforce 5MB per image, set server disk quota, plan migration to S3-compatible storage in Sprint 3/4 |
| R-03 | Full-text search performance degrades at scale | Medium | High | Use PostgreSQL tsvector indexes from Sprint 3. Monitor query times. Plan Elasticsearch migration if needed. |
| R-04 | Explore feed algorithm is too slow at scale | Medium | High | Materialize feed ranking scores on write, not on read. Cache feed pages in Redis (Sprint 3+). |
| R-05 | Email deliverability (verification + reset emails land in spam) | Medium | High | Use a reputable transactional email provider (SendGrid/Mailgun). Set SPF/DKIM records. |
| R-06 | Google OAuth token mismanagement (security) | Low | Critical | Store only user email + Google ID. Never store Google access tokens. Validate tokens server-side. |
| R-07 | React Native app diverges from web feature parity | High | Medium | Share API client code between web and mobile. Define API contracts in Sprint 1, not Sprint 4. |
| R-08 | Report abuse (users spamming reports) | Medium | Medium | One report per user per content item enforced at DB level. Rate limit report endpoint. |
| R-09 | TDD coverage < 80% slows pipeline | Medium | Medium | Developer writes tests for every AC before implementation. QA blocks Gate 4 automatically. |
| R-10 | Scope creep in Sprint 1 (too many features) | Medium | High | Strictly follow MoSCoW. Only M-01 to M-14 in Sprint 1. PM enforces scope at every gate. |
