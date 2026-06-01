# Database Schema — Blog Engine
# Version: 1.0 — 2026-05-30

All tables use UUID primary keys. PostgreSQL 15+.

---

## users
```sql
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(50) UNIQUE NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),                    -- NULL for Google OAuth users
    google_id     VARCHAR(255) UNIQUE,             -- NULL for email users
    role          VARCHAR(20) NOT NULL DEFAULT 'user',  -- guest|user|moderator|admin|owner
    verified      BOOLEAN NOT NULL DEFAULT FALSE,
    avatar_url    VARCHAR(500),
    bio           TEXT,
    favorite_quote TEXT,
    locked_until  TIMESTAMPTZ,                     -- brute force lockout
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
```

## email_verifications
```sql
CREATE TABLE email_verifications (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## password_resets
```sql
CREATE TABLE password_resets (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## refresh_tokens
```sql
CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## categories
```sql
CREATE TABLE categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(100) UNIQUE NOT NULL,
    slug       VARCHAR(100) UNIQUE NOT NULL,
    created_by UUID REFERENCES users(id),         -- NULL = predefined by system
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## blogs
```sql
CREATE TABLE blogs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id     UUID NOT NULL REFERENCES users(id),
    title         VARCHAR(500) NOT NULL,
    content       TEXT NOT NULL,                  -- sanitized HTML
    excerpt       VARCHAR(300),                   -- auto-generated first ~100 chars
    thumbnail_url VARCHAR(500),
    privacy       VARCHAR(20) NOT NULL DEFAULT 'public', -- public|friend_only|only_me
    status        VARCHAR(20) NOT NULL DEFAULT 'draft',  -- draft|published
    like_count    INT NOT NULL DEFAULT 0,
    dislike_count INT NOT NULL DEFAULT 0,
    comment_count INT NOT NULL DEFAULT 0,
    read_time_min INT NOT NULL DEFAULT 1,         -- estimated minutes
    feed_score    FLOAT NOT NULL DEFAULT 0,       -- denormalized for feed ordering (ADR-006)
    search_vector TSVECTOR,                       -- full-text search (ADR-002)
    published_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_blogs_author ON blogs(author_id);
CREATE INDEX idx_blogs_feed_score ON blogs(feed_score DESC) WHERE status = 'published' AND privacy = 'public';
CREATE INDEX idx_blogs_published_at ON blogs(published_at DESC);
CREATE INDEX idx_blogs_search ON blogs USING GIN(search_vector);
```

## blog_tags
```sql
CREATE TABLE tags (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL
);

CREATE TABLE blog_tags (
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    tag_id  UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (blog_id, tag_id)
);
CREATE INDEX idx_blog_tags_tag ON blog_tags(tag_id);
```

## blog_categories
```sql
CREATE TABLE blog_categories (
    blog_id     UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (blog_id, category_id)
);
```

## images
```sql
CREATE TABLE images (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id    UUID REFERENCES blogs(id) ON DELETE CASCADE, -- NULL = profile avatar
    user_id    UUID NOT NULL REFERENCES users(id),
    filename   VARCHAR(255) NOT NULL,
    r2_key     VARCHAR(500) NOT NULL,   -- object key in Cloudflare R2 bucket
    url        VARCHAR(500) NOT NULL,   -- public R2 URL (served via Cloudflare CDN)
    size_bytes INT NOT NULL,
    mime_type  VARCHAR(50) NOT NULL,    -- image/jpeg | image/png | image/webp
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## follows
```sql
CREATE TABLE follows (
    follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, followee_id),
    CHECK (follower_id != followee_id)
);
CREATE INDEX idx_follows_followee ON follows(followee_id);
```

## friend_requests
```sql
CREATE TABLE friend_requests (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending|accepted|rejected
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (sender_id, receiver_id),
    CHECK (sender_id != receiver_id)
);
```

## friends
```sql
CREATE TABLE friends (
    user_id_1  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_id_2  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id_1, user_id_2),
    CHECK (user_id_1 < user_id_2)  -- enforce canonical ordering
);
CREATE INDEX idx_friends_user2 ON friends(user_id_2);
```

## blocks
```sql
CREATE TABLE blocks (
    blocker_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (blocker_id, blocked_id),
    CHECK (blocker_id != blocked_id)
);
CREATE INDEX idx_blocks_blocked ON blocks(blocked_id);
```

## reactions
```sql
CREATE TABLE reactions (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blog_id    UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    type       VARCHAR(10) NOT NULL,  -- like|dislike
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, blog_id)
);
```

## comments
```sql
CREATE TABLE comments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id    UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    author_id  UUID NOT NULL REFERENCES users(id),
    parent_id  UUID REFERENCES comments(id) ON DELETE CASCADE, -- NULL = top-level
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_comments_blog ON comments(blog_id);
CREATE INDEX idx_comments_parent ON comments(parent_id);
```

## notifications
```sql
CREATE TABLE notifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        VARCHAR(50) NOT NULL,
    -- like_blog|dislike_blog|comment_blog|reply_comment|new_follower
    -- friend_request|friend_accepted|content_reported
    actor_id    UUID REFERENCES users(id),        -- who triggered it
    blog_id     UUID REFERENCES blogs(id) ON DELETE CASCADE,
    comment_id  UUID REFERENCES comments(id) ON DELETE CASCADE,
    read        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_notifications_user ON notifications(user_id, read, created_at DESC);
```

## reports
```sql
CREATE TABLE reports (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id  UUID NOT NULL REFERENCES users(id),
    blog_id      UUID REFERENCES blogs(id) ON DELETE CASCADE,
    comment_id   UUID REFERENCES comments(id) ON DELETE CASCADE,
    reason       VARCHAR(50) NOT NULL,
    -- spam|inappropriate|harassment|misinformation|other
    status       VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending|resolved
    resolved_by  UUID REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (reporter_id, blog_id),      -- one report per user per blog
    UNIQUE (reporter_id, comment_id),   -- one report per user per comment
    CHECK (
        (blog_id IS NOT NULL AND comment_id IS NULL) OR
        (blog_id IS NULL AND comment_id IS NOT NULL)
    )
);
```
