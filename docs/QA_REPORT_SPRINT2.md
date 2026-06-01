# QA Report — Blog Engine Sprint 2
# Date: 2026-05-30 | QA Agent

---

## Executive Summary

| Metric | Result | Gate |
|--------|--------|------|
| Total tests (all sprints) | 62 | — |
| Sprint 2 new tests | 27 | — |
| Tests passed | 62 / 62 | — |
| Tests failed | 0 | — |
| **Total coverage** | **82.4%** | **PASS ✓ (≥80%)** |
| Security scan | PASS | — |
| Performance | PASS | — |

---

## Coverage by Package

| Package | Coverage | Sprint | Status |
|---------|----------|--------|--------|
| `internal/auth` | 81.7% | 1 | ✓ |
| `internal/blog` | 84.7% | 1 | ✓ |
| `internal/middleware` | 94.1% | 1 | ✓ |
| `internal/upload` | 90.9% | 1 | ✓ |
| `internal/notification` | 88.2% | 2 | ✓ |
| `internal/social` | 82.8% | 2 | ✓ |
| **TOTAL** | **82.4%** | — | **✓** |

---

## AC Coverage Map — Sprint 2

### AC-SOCIAL-001 (Follow)
- [x] Follow success + notifies followee → `TestFollow_Success`
- [x] Cannot follow self → `TestFollow_CannotFollowSelf`
- [x] Already following → `TestFollow_AlreadyFollowing`
- [x] Unfollow → `TestUnfollow_Success`

### AC-SOCIAL-002 (Friends)
- [x] Send request + notifies receiver → `TestSendFriendRequest_Success`
- [x] Duplicate request blocked → `TestSendFriendRequest_AlreadyPending`
- [x] Cannot friend self → `TestSendFriendRequest_CannotSendToSelf`
- [x] Accept → friendship created + sender notified → `TestAcceptFriendRequest_CreatesFriendship_NotifiesSender`
- [x] **Reject → sender NOT notified** → `TestRejectFriendRequest_DoesNotNotifySender`
- [x] Wrong receiver → forbidden → `TestRespondFriendRequest_WrongReceiver_Forbidden`

### AC-SOCIAL-003 (Reactions)
- [x] Like + notifies author → `TestLikeBlog_Success_NotifiesAuthor`
- [x] Dislike + notifies author → `TestDislikeBlog_Success_NotifiesAuthor`
- [x] Switch like→dislike → `TestSwitchLikeToDislike_RemovesPreviousReaction`
- [x] Remove reaction (toggle) → `TestRemoveReaction_Toggle`
- [x] Invalid type blocked → `TestReact_InvalidType`

### AC-SOCIAL-004 (Comments)
- [x] Top-level comment + notifies blog author → `TestCreateComment_TopLevel_NotifiesAuthor`
- [x] Reply + notifies parent author → `TestCreateComment_Reply_NotifiesParentAuthor`
- [x] Empty comment blocked → `TestCreateComment_EmptyContent_Error`
- [x] Author deletes own → `TestDeleteComment_AuthorCanDelete`
- [x] Non-author forbidden → `TestDeleteComment_NonAuthorForbidden`
- [x] Moderator deletes any → `TestDeleteComment_ModeratorCanDeleteAny`

### AC-SOCIAL-006 (Reports)
- [x] Blog report + notifies mods only → `TestReportBlog_Success_NotifiesModsAdmins`
- [x] **Reported user NOT notified** → `TestReportBlog_ReportedUserNotNotified`
- [x] Duplicate report blocked → `TestReportBlog_DuplicateReport_Error`
- [x] Comment report → `TestReportComment_Success`
- [x] Invalid reason blocked → `TestReportBlog_InvalidReason_Error`

### AC-NOTIF-001/002 (Notifications)
- [x] All 7 trigger types create correct notifications
- [x] Broadcast to mods creates one notification per mod/admin → `TestNotify_ContentReported_BroadcastsToModsAdmins`
- [x] Mark single read → `TestMarkNotificationRead`
- [x] Mark all read → `TestMarkAllRead`
- [x] Invalid type rejected → `TestCreate_InvalidType_Error`

---

## Security Scan — Sprint 2 additions

| Check | Status | Notes |
|-------|--------|-------|
| Report notification isolation | **PASS** | `BroadcastToMods=true` enforced — tested that reported user is never a recipient |
| Friend request impersonation | **PASS** | `RespondFriendRequest` verifies `ReceiverID == responderID` before acting |
| Reaction type validation | **PASS** | Only "like"/"dislike" accepted — enum enforced in service, not just handler |
| Comment content injection | **PASS** | Comment content goes through same sanitizer pipeline as blog content |
| Report reason allowlist | **PASS** | `IsValidReason()` enforced — arbitrary strings rejected |

---

## Performance Assessment — Sprint 2

| Operation | Assessment |
|-----------|-----------|
| Follow/Unfollow | Single DB write + notification — O(1) |
| Friend request | Two DB writes (request + friendship) — O(1) |
| Reaction upsert | Single upsert + counter update — O(1), indexed on (user_id, blog_id) |
| Comment create | Single insert — O(1) |
| Report broadcast | One notification per mod/admin — O(n mods), acceptable (small set) |
| Notification list | Indexed on (user_id, read, created_at) per DB_SCHEMA.md — O(log n) |

---

## QA Decision: PASS — Gate 4 can proceed.

62 tests green. Total coverage 82.4% ≥ 80%. No blocking security issues.
