CREATE TABLE contents (
    content_id VARCHAR(36) PRIMARY KEY,
    author_user_id VARCHAR(36) NOT NULL,
    text TEXT NOT NULL,
    media_urls JSONB NOT NULL DEFAULT '[]',
    status SMALLINT NOT NULL,
    like_count BIGINT NOT NULL DEFAULT 0,
    comment_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL,
    published_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL
);
CREATE TABLE content_likes (
    content_id VARCHAR(36) NOT NULL REFERENCES contents(content_id) ON DELETE CASCADE,
    user_id VARCHAR(36) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (content_id, user_id)
);
CREATE TABLE content_comments (
    comment_id VARCHAR(36) PRIMARY KEY,
    content_id VARCHAR(36) NOT NULL REFERENCES contents(content_id) ON DELETE CASCADE,
    author_user_id VARCHAR(36) NOT NULL,
    parent_id VARCHAR(36) NOT NULL DEFAULT '',
    text TEXT NOT NULL,
    status SMALLINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
