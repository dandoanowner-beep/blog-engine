# Acceptance Criteria — Blog Engine
# Version: 1.0 — 2026-05-30

---

## AC-AUTH-001: Email Registration
- [ ] User can register with email, username, password (min 8 chars)
- [ ] Duplicate email returns 400 with clear error message
- [ ] Verification email sent within 30 seconds of registration
- [ ] Account created with status `unverified`
- [ ] Unverified user cannot publish blogs (returns 403)

## AC-AUTH-002: Google OAuth
- [ ] Google OAuth button redirects to Google consent screen
- [ ] On success, user account created or logged into existing account
- [ ] Google users are automatically `verified`
- [ ] Google users can publish immediately after signup

## AC-AUTH-003: Email Verification
- [ ] Clicking valid link within 24h sets status to `verified`
- [ ] Expired link shows error and prompts for new email
- [ ] New verification email can be requested once per 5 minutes
- [ ] Already-verified link shows "already verified" without error

## AC-AUTH-004: Login
- [ ] Correct credentials return JWT access token + refresh token
- [ ] Wrong credentials return 401 with "Invalid email or password"
- [ ] JWT expires after 15 minutes; refresh token valid 7 days
- [ ] After 5 failed attempts, account locked for 15 minutes

## AC-AUTH-005: Password Reset
- [ ] Reset email sent within 1 minute of request
- [ ] Reset link expires after 1 hour
- [ ] Using expired link shows error message
- [ ] Password successfully reset with new password (min 8 chars)

## AC-AUTH-006: Block System
- [ ] Blocked user's blogs absent from all feeds of the blocker
- [ ] Blocker's profile returns 404 (or blank) to the blocked user
- [ ] Block is silent — no notification sent to blocked user
- [ ] Blocker can unblock from their settings page

---

## AC-BLOG-001: Create Blog
- [ ] Editor supports bold, italic, underline, headings (H1-H4), lists, blockquote, divider
- [ ] Markdown rendered correctly in preview
- [ ] Code blocks display with syntax highlighting
- [ ] Thumbnail image upload succeeds for files ≤ 5MB (JPEG, PNG, WEBP)
- [ ] Upload rejected with error for files > 5MB
- [ ] Multiple inline images can be inserted in body (each ≤ 5MB)
- [ ] At least one tag required — validation error if missing
- [ ] At least one category required — validation error if missing
- [ ] Privacy selector shows three options: Public / Friend-only / Only-me
- [ ] Published blog appears in Explore feed within 60 seconds

## AC-BLOG-002: Draft
- [ ] Draft saved with a single click, no publish action triggered
- [ ] Draft visible only in author's own dashboard/profile
- [ ] Draft not visible in Explore or Following feed
- [ ] Draft can be edited and published at any time
- [ ] Published draft transitions to correct privacy setting

## AC-BLOG-003: Privacy
- [ ] Public blog readable (partial) by guest users
- [ ] Guest sees top ~30% content; remainder blurred with signup prompt
- [ ] Friend-only blog not visible in feeds or search for non-friends
- [ ] Only-me blog not visible to anyone other than the author
- [ ] Privacy mode can be changed after publishing

## AC-BLOG-004: Delete Blog
- [ ] Author can delete own blog with confirmation dialog
- [ ] On delete: blog + all comments + all reactions removed from DB
- [ ] Moderator / Admin / Owner can delete any blog
- [ ] Deleted blog URL returns 404

---

## AC-FEED-001: Blog Card
- [ ] Card displays: thumbnail, title, excerpt (~100 chars), author avatar + name, read time, 2-3 tags, like count, dislike count, comment count
- [ ] 3 cards displayed per row on desktop
- [ ] Cards are responsive on mobile

## AC-FEED-002: Explore Feed
- [ ] All public blogs shown to all signed-in users
- [ ] Blogs from followed writers ranked higher
- [ ] Algorithm factors: recency + engagement score (likes + comments)
- [ ] Pagination: numbered pages, 12 blogs per page
- [ ] Filter by tag returns only blogs with that tag
- [ ] Filter by category returns only blogs in that category

## AC-FEED-003: Following Feed
- [ ] Only shows blogs from users the current user follows
- [ ] Ordered newest first
- [ ] Empty state shown when user follows nobody
- [ ] Paginated (numbered pages, 12 per page)

---

## AC-SOCIAL-001: Follow
- [ ] Follow button on user profile works for signed-in users only
- [ ] Follower count increments on target user's profile
- [ ] Following count increments on current user's profile
- [ ] Target user receives in-app notification
- [ ] Unfollow removes from Following feed immediately

## AC-SOCIAL-002: Friend Request
- [ ] Send request: request stored as `pending`, recipient notified
- [ ] Accept: both users get `friends` status, sender notified
- [ ] Reject: request removed, sender NOT notified
- [ ] Cannot send duplicate request while one is pending
- [ ] Friend-only blogs visible to mutual friends after acceptance
- [ ] Unfriend removes friend status and friend-only blog access

## AC-SOCIAL-003: Like / Dislike
- [ ] Signed-in user can like a blog (one like per user per blog)
- [ ] Signed-in user can dislike a blog (one dislike per user)
- [ ] Cannot like AND dislike simultaneously — switching removes the previous
- [ ] Clicking active reaction again removes it (toggle off)
- [ ] Author notified on like and on dislike separately
- [ ] Counts update in real-time (or near real-time within 5 seconds)

## AC-SOCIAL-004: Comments
- [ ] Signed-in user can post a comment on any readable blog
- [ ] Comment appears without page reload
- [ ] Reply nested under parent comment (threaded display)
- [ ] No limit on reply depth
- [ ] Author notified on new comment; comment author notified on reply
- [ ] Author can delete own comment; Moderator/Admin can delete any comment
- [ ] Deleting a comment removes all its nested replies

## AC-SOCIAL-005: Sharing
- [ ] "Copy link" copies the blog's full URL to clipboard with confirmation toast
- [ ] "Share to Facebook" opens Facebook share dialog with blog URL pre-filled
- [ ] "Share to Zalo" opens Zalo share with blog URL
- [ ] Share buttons available on blog detail page

## AC-SOCIAL-006: Report
- [ ] Report option available on blogs and comments for signed-in users
- [ ] Reporter must select a reason from predefined list
- [ ] Reporter sees "Report submitted" confirmation
- [ ] Reported user receives NO notification
- [ ] All Moderators and Admins receive in-app notification with link to reported content
- [ ] Same user cannot report the same content twice

---

## AC-NOTIF-001: Notifications
- [ ] Bell icon in header shows unread notification count badge
- [ ] Clicking bell opens notification panel (newest first)
- [ ] Each notification has text, timestamp, and link to relevant content
- [ ] Clicking notification marks it as read and navigates to content
- [ ] "Mark all as read" button clears all unread badges
- [ ] All 7 user-facing trigger events produce correct notifications
- [ ] Report notifications delivered only to Moderators and Admins

---

## AC-SEARCH-001: Search
- [ ] Search bar accessible on all pages
- [ ] Results grouped into: Blogs, Users, Tags sections
- [ ] Matches found in: blog title, blog content (full-text), tags, categories, usernames
- [ ] Friend-only blogs not shown to non-friends in search results
- [ ] Only-me blogs never appear in search results (except to author)
- [ ] Search results paginated (numbered pages, 10 results per section)
- [ ] Empty state shown when no results found

---

## AC-PROFILE-001: User Profile
- [ ] Profile page shows: avatar, username, bio, favorite quote, blogs grid, follower/following/friend counts
- [ ] Only published public blogs visible to guests
- [ ] Published public + friend-only blogs visible to mutual friends
- [ ] Drafts and only-me blogs visible only to profile owner
- [ ] Edit profile: update avatar (≤ 5MB image), bio, favorite quote, username
- [ ] Username change must be unique across platform

---

## AC-ADMIN-001: Admin Dashboard
- [ ] Dashboard accessible only to Admin and Owner roles (others get 403)
- [ ] User list with role badges and ability to change roles
- [ ] Admin can promote User to Moderator
- [ ] Reports queue lists all unresolved reports with: content preview, reason, reporter username, timestamp
- [ ] Admin/Owner can delete reported content directly from queue
- [ ] Platform stats visible: total users, total blogs, total comments, new signups this week
