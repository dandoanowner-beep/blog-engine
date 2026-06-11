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

---

## i18n Sprint Risks (2026-06-07)

| ID | Risk | Probability | Impact | Mitigation |
|----|------|-------------|--------|------------|
| R-I18N-01 | Claude API latency blocks blog publish response | Medium | High | Run translation in a goroutine after blog is saved — API response returns immediately; translation happens in background |
| R-I18N-02 | Translation quality poor for specialized Vietnamese content | Low | Low | claude-sonnet-4-6 handles VI/EN well; author can re-save blog to trigger re-translation |
| R-I18N-03 | UI strings missed — t() key falls back to raw key string (bad UX) | Medium | Medium | Systematically extract all strings before replacing; test that all keys exist in vi.json + en.json |
| R-I18N-04 | Existing blogs have no English translation after migration | Certain | Low | Acceptable — fallback to VI content is specified in FRD FR-I18N-005; no backfill required |
| R-I18N-05 | Claude API quota exceeded during testing / bulk blog creation | Low | Medium | Personal blog volume is very low; graceful failure (translation_status=failed) prevents data loss |
| R-I18N-06 | react-i18next adds bundle size / complexity to frontend | Low | Low | Industry-standard library; tree-shaking keeps bundle impact minimal |
