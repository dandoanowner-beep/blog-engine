-- Blog Engine — Initial Schema
-- Sprint 1-3 complete

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(50)  UNIQUE NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    google_id     VARCHAR(255) UNIQUE,
    role          VARCHAR(20)  NOT NULL DEFAULT 'user',
    verified      BOOLEAN      NOT NULL DEFAULT FALSE,
    avatar_url    VARCHAR(500),
    bio           TEXT,
    favorite_quote TEXT,
    login_attempts INT         NOT NULL DEFAULT 0,
    locked_until  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_users_email    ON users(email);
CREATE INDEX idx_users_username ON users(username);

CREATE TABLE email_verifications (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used       BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE password_resets (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used       BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE refresh_tokens (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE categories (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(100) UNIQUE NOT NULL,
    slug       VARCHAR(100) UNIQUE NOT NULL,
    created_by UUID        REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE blogs (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id     UUID        NOT NULL REFERENCES users(id),
    title         VARCHAR(500) NOT NULL,
    content       TEXT        NOT NULL,
    excerpt       VARCHAR(300),
    thumbnail_url VARCHAR(500),
    privacy       VARCHAR(20) NOT NULL DEFAULT 'public',
    status        VARCHAR(20) NOT NULL DEFAULT 'draft',
    like_count    INT         NOT NULL DEFAULT 0,
    dislike_count INT         NOT NULL DEFAULT 0,
    comment_count INT         NOT NULL DEFAULT 0,
    read_time_min INT         NOT NULL DEFAULT 1,
    feed_score    FLOAT       NOT NULL DEFAULT 0,
    search_vector TSVECTOR,
    published_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_blogs_author       ON blogs(author_id);
CREATE INDEX idx_blogs_feed_score   ON blogs(feed_score DESC) WHERE status='published' AND privacy='public';
CREATE INDEX idx_blogs_published_at ON blogs(published_at DESC);
CREATE INDEX idx_blogs_search       ON blogs USING GIN(search_vector);

CREATE TABLE tags (
    id   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL
);

CREATE TABLE blog_tags (
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    tag_id  UUID NOT NULL REFERENCES tags(id)  ON DELETE CASCADE,
    PRIMARY KEY (blog_id, tag_id)
);
CREATE INDEX idx_blog_tags_tag ON blog_tags(tag_id);

CREATE TABLE blog_categories (
    blog_id     UUID NOT NULL REFERENCES blogs(id)       ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id)  ON DELETE CASCADE,
    PRIMARY KEY (blog_id, category_id)
);

CREATE TABLE images (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id    UUID        REFERENCES blogs(id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(id),
    filename   VARCHAR(255) NOT NULL,
    r2_key     VARCHAR(500) NOT NULL,
    url        VARCHAR(500) NOT NULL,
    size_bytes INT         NOT NULL,
    mime_type  VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE blocks (
    blocker_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (blocker_id, blocked_id),
    CHECK (blocker_id != blocked_id)
);
CREATE INDEX idx_blocks_blocked ON blocks(blocked_id);

CREATE TABLE follows (
    follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, followee_id),
    CHECK (follower_id != followee_id)
);
CREATE INDEX idx_follows_followee ON follows(followee_id);

CREATE TABLE friend_requests (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (sender_id, receiver_id),
    CHECK (sender_id != receiver_id)
);

CREATE TABLE friends (
    user_id_1  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_id_2  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id_1, user_id_2),
    CHECK (user_id_1 < user_id_2)
);
CREATE INDEX idx_friends_user2 ON friends(user_id_2);

CREATE TABLE reactions (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blog_id    UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    type       VARCHAR(10) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, blog_id)
);

CREATE TABLE comments (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id    UUID        NOT NULL REFERENCES blogs(id)     ON DELETE CASCADE,
    author_id  UUID        NOT NULL REFERENCES users(id),
    parent_id  UUID        REFERENCES comments(id)           ON DELETE CASCADE,
    content    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_comments_blog   ON comments(blog_id);
CREATE INDEX idx_comments_parent ON comments(parent_id);

CREATE TABLE notifications (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       VARCHAR(50) NOT NULL,
    actor_id   UUID        REFERENCES users(id),
    blog_id    UUID        REFERENCES blogs(id)    ON DELETE CASCADE,
    comment_id UUID        REFERENCES comments(id) ON DELETE CASCADE,
    read       BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_notifications_user ON notifications(user_id, read, created_at DESC);

CREATE TABLE reports (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id  UUID        NOT NULL REFERENCES users(id),
    blog_id      UUID        REFERENCES blogs(id)    ON DELETE CASCADE,
    comment_id   UUID        REFERENCES comments(id) ON DELETE CASCADE,
    reason       VARCHAR(50) NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'pending',
    resolved_by  UUID        REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (reporter_id, blog_id),
    UNIQUE (reporter_id, comment_id),
    CHECK (
        (blog_id IS NOT NULL AND comment_id IS NULL) OR
        (blog_id IS NULL AND comment_id IS NOT NULL)
    )
);

-- Trigger: auto-update search_vector on blogs insert/update
CREATE OR REPLACE FUNCTION update_blog_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector := to_tsvector('english', COALESCE(NEW.title,'') || ' ' || COALESCE(NEW.content,''));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER blogs_search_vector_update
    BEFORE INSERT OR UPDATE ON blogs
    FOR EACH ROW EXECUTE FUNCTION update_blog_search_vector();
