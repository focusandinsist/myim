CREATE TABLE follows (
    follower_user_id VARCHAR(36) NOT NULL,
    followee_user_id VARCHAR(36) NOT NULL,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_user_id, followee_user_id),
    CHECK (follower_user_id <> followee_user_id)
);
CREATE INDEX follows_follower_status_updated_idx ON follows (follower_user_id, status, updated_at DESC);
CREATE INDEX follows_followee_status_updated_idx ON follows (followee_user_id, status, updated_at DESC);
