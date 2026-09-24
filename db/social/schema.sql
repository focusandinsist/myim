-- users is owned by user-service but is visible because services currently share PostgreSQL.
CREATE TABLE users (
    user_id VARCHAR(36) PRIMARY KEY,
    status SMALLINT NOT NULL DEFAULT 0
);

CREATE TABLE follows (
    follower_user_id VARCHAR(36) NOT NULL,
    followee_user_id VARCHAR(36) NOT NULL,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_user_id, followee_user_id),
    CHECK (follower_user_id <> followee_user_id)
);
