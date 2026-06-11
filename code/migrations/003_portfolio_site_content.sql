-- CR-002: Portfolio projects + owner-editable site content (Author page)

CREATE TABLE projects (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    title         VARCHAR(200) NOT NULL,
    description   TEXT         NOT NULL DEFAULT '',
    tech_stack    VARCHAR(500) NOT NULL DEFAULT '',
    repo_url      VARCHAR(500) NOT NULL DEFAULT '',
    demo_url      VARCHAR(500) NOT NULL DEFAULT '',
    thumbnail_url VARCHAR(500) NOT NULL DEFAULT '',
    sort_order    INT          NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_projects_sort ON projects(sort_order, created_at DESC);

-- Single-document storage for owner-authored site pages (key = 'about' for now)
CREATE TABLE site_content (
    key        VARCHAR(50)  PRIMARY KEY,
    content    TEXT         NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
