# Task Roadmap — Blog Engine
# Version: 1.0 — 2026-05-30

---

## MoSCoW Backlog

### MUST HAVE (Sprint 1 — Core Foundation)
| ID | Feature | Effort |
|----|---------|--------|
| M-01 | User registration (email + password) | M |
| M-02 | Email verification flow | M |
| M-03 | Google OAuth login / registration | M |
| M-04 | JWT login + refresh token + password reset | M |
| M-05 | User roles: Guest, User, Moderator, Admin, Owner | M |
| M-06 | Blog creation: WYSIWYG editor + markdown + code blocks | L |
| M-07 | Blog thumbnail upload (5MB limit) + inline images | M |
| M-08 | Tags + Categories (predefined + user-created) | M |
| M-09 | Blog privacy modes: Public / Friend-only / Only-me | M |
| M-10 | Draft system (save + publish later) | S |
| M-11 | Blog card component (3 per row, all metadata) | M |
| M-12 | Explore feed (algorithmic ranking, paginated) | L |
| M-13 | Following feed (newest first, paginated) | M |
| M-14 | Guest partial read + signup prompt | S |

### MUST HAVE (Sprint 2 — Social Core)
| ID | Feature | Effort |
|----|---------|--------|
| M-15 | Follow / Unfollow users | S |
| M-16 | Friend request system (send / accept / reject) | M |
| M-17 | Like / Dislike reactions | S |
| M-18 | Threaded comments | M |
| M-19 | In-app notifications (all 7 triggers) | L |
| M-20 | Block / Unblock system | S |
| M-21 | Report blog / comment → notify Moderators + Admins | M |
| M-22 | Delete blog (author + Moderator/Admin) | S |
| M-23 | Delete comment (author + Moderator/Admin) | S |

### MUST HAVE (Sprint 3 — Search + Profile + Admin)
| ID | Feature | Effort |
|----|---------|--------|
| M-24 | Universal search (blogs, users, tags, categories — full-text) | L |
| M-25 | User profile page (all fields + blog grid) | M |
| M-26 | Profile editing (avatar, bio, favorite quote, username) | S |
| M-27 | Admin dashboard (user management + reports queue + stats) | L |
| M-28 | Share to Facebook + Zalo + copy link | S |

### SHOULD HAVE (Sprint 4)
| ID | Feature | Effort |
|----|---------|--------|
| S-01 | React Native mobile app (iOS + Android) | XL |
| S-02 | Explore feed filter by tag/category (UI refinement) | S |
| S-03 | Notification mark-all-as-read | S |

### COULD HAVE (Future)
| ID | Feature |
|----|---------|
| C-01 | Reading history ("blogs you've read") |
| C-02 | Bookmarks / saved blogs |
| C-03 | Blog series (multi-part posts) |
| C-04 | Email digest notifications |

### WON'T HAVE (This Version)
- Real-time features (WebSocket live updates)
- Paid subscriptions / paywalls
- Video content in blogs
- Account deletion

---

## Sprint Summary

| Sprint | Focus | Key Deliverables |
|--------|-------|-----------------|
| Sprint 1 | Core Foundation | Auth, blog CRUD, feed, guest experience |
| Sprint 2 | Social Core | Follow, friends, reactions, comments, notifications, block, report |
| Sprint 3 | Discovery + Admin | Search, profile, admin dashboard, sharing |
| Sprint 4 | Mobile | React Native app |
