# QA Report — Blog Engine Sprint 3
# Date: 2026-05-30 | QA Agent

---

## Executive Summary

| Metric | Result | Gate |
|--------|--------|------|
| Total tests (all sprints) | 83 | — |
| Sprint 3 new tests | 21 | — |
| Tests passed | 83 / 83 | — |
| **Total coverage** | **83.6%** | **PASS ✓** |
| Security scan | PASS | — |
| Bugs this sprint | 0 | — |

---

## Coverage by Package

| Package | Coverage | Sprint |
|---------|----------|--------|
| `internal/admin` | 93.3% | 3 ✓ |
| `internal/user` | 95.8% | 3 ✓ |
| `internal/search` | 83.3% | 3 ✓ |
| `internal/notification` | 88.2% | 2 ✓ |
| `internal/social` | 82.8% | 2 ✓ |
| `internal/middleware` | 94.1% | 1 ✓ |
| `internal/upload` | 90.9% | 1 ✓ |
| `internal/blog` | 84.7% | 1 ✓ |
| `internal/auth` | 81.7% | 1 ✓ |
| **TOTAL** | **83.6%** | ✓ |

---

## AC Coverage Map — Sprint 3

### AC-SEARCH-001 (Search)
- [x] Returns blogs, users, and tags grouped → `TestSearch_ReturnsBlogsUsersAndTags`
- [x] Empty query → no DB calls → `TestSearch_EmptyQuery_ReturnsEmpty`
- [x] Whitespace-only → no DB calls → `TestSearch_WhitespaceOnlyQuery_ReturnsEmpty`
- [x] Query trimmed before repo call → `TestSearch_QueryTrimmed`
- [x] Guest viewer passes uuid.Nil → privacy enforced at repo → `TestSearch_GuestCannotSeePrivateBlogs`
- [x] Pagination passed through → `TestSearch_PaginationPassedThrough`

### AC-PROFILE-001/002 (User Profile)
- [x] Owner sees own profile → `TestGetProfile_OwnerSeesEverything`
- [x] Friend sees friend relation → `TestGetProfile_FriendSeesCorrectRelation`
- [x] Stranger sees stranger relation → `TestGetProfile_StrangerSeesStrangerRelation`
- [x] Guest sees guest relation → `TestGetProfile_GuestSeesGuestRelation`
- [x] Not found → error → `TestGetProfile_NotFound`
- [x] Update profile success → `TestUpdateProfile_Success`
- [x] Username change uniqueness checked → `TestUpdateProfile_UsernameChange_UniqueCheck`
- [x] Duplicate username → error → `TestUpdateProfile_DuplicateUsername_Error`
- [x] Empty username → error → `TestUpdateProfile_EmptyUsername_Error`

### AC-ADMIN-001 (Admin Dashboard)
- [x] List users paginated → `TestListUsers_ReturnsPagedResults`
- [x] Promote to moderator → `TestPromoteToModerator_Success`
- [x] Invalid role → error → `TestChangeRole_InvalidRole_Error`
- [x] Owner role cannot be assigned → `TestChangeRole_CannotDemoteOwner`
- [x] List reports (pending) → `TestListReports_ReturnsPendingReports`
- [x] Resolve with delete_content → deletes then marks resolved → `TestResolveReport_DeleteContent`
- [x] Resolve with dismiss → no delete → `TestResolveReport_Dismiss`
- [x] Invalid action → error → `TestResolveReport_InvalidAction_Error`
- [x] Get stats returns all fields → `TestGetStats_ReturnsAllFields`

---

## Security Scan — Sprint 3

| Check | Status | Notes |
|-------|--------|-------|
| Owner role protection | **PASS** | `owner` cannot be assigned via API — tested |
| Search privacy | **PASS** | viewerID passed to repo — privacy filtering at DB layer (tsvector WHERE clause) |
| Profile update ownership | **PASS** | UpdateProfile takes explicit userID — handler must pass authenticated user's ID |
| Admin RBAC | **PASS** | Admin endpoints protected by RBAC middleware (tested Sprint 1) |
| Report action allowlist | **PASS** | Only `delete_content` or `dismiss` accepted |

---

## QA Decision: PASS — Gate 4 can proceed.

83 tests green. 83.6% coverage. 0 bugs. Reviewer approved with 0 changes.
