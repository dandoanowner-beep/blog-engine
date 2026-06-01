# Definition of Done — Blog Engine
# Version: 1.0 — 2026-05-30

A feature is DONE only when ALL of the following are true.

---

## Code Quality
- [ ] Code compiles with zero errors and zero warnings
- [ ] All acceptance criteria for the feature are met and verified
- [ ] Code follows Go standard conventions (backend) and React best practices (frontend)
- [ ] No hardcoded secrets, API keys, or credentials in code
- [ ] All inputs validated and sanitized (OWASP Top 10 compliance)

## Testing
- [ ] Unit tests written for all business logic functions
- [ ] Integration tests written for all API endpoints
- [ ] Test coverage ≥ 80% (enforced — pipeline blocks below this)
- [ ] All tests pass with zero failures (`go test ./...` exits 0)
- [ ] Edge cases tested: empty inputs, max size limits, unauthorized access attempts

## Documentation
- [ ] API endpoint documented in API_CONTRACT.md (OpenAPI format)
- [ ] Any new architectural decision recorded as an ADR
- [ ] Any bug found during development logged in BUGS.md
- [ ] Any bug fixed has status updated to FIXED with lesson learned

## Security
- [ ] All routes requiring authentication protected with JWT middleware
- [ ] Role-based access control enforced on protected endpoints
- [ ] SQL queries use parameterized statements (no raw string concatenation)
- [ ] File uploads validated: type (JPEG/PNG/WEBP only) + size (≤ 5MB)

## Review
- [ ] Code reviewed by Reviewer Agent — decision is APPROVE
- [ ] No REQUEST_CHANGES items left unresolved

## Sprint Done (end of sprint)
- [ ] All Must-have items for the sprint are in DONE state
- [ ] QA Agent run completed with coverage ≥ 80%
- [ ] Gate 4 approved by human (Owner)
- [ ] DevOps Agent produced CI/CD pipeline, IaC, and rollback procedure
- [ ] PM Agent sprint review completed: SPRINT_LOG.md updated, User_Guideline.md updated
- [ ] PIPELINE_STATUS.md updated to reflect completed sprint
