-- name: CreateContent :exec
INSERT INTO contents (content_id, author_user_id, text, media_urls, status, like_count, comment_count, created_at, published_at, updated_at)
VALUES ($1, $2, $3, $4, $5, 0, 0, $6, $7, $8);

-- name: GetContent :one
SELECT content_id, author_user_id, text, media_urls, status, like_count, comment_count, created_at, published_at, updated_at
FROM contents WHERE content_id = $1;

-- name: ListContents :many
SELECT content_id, author_user_id, text, media_urls, status, like_count, comment_count, created_at, published_at, updated_at
FROM contents WHERE author_user_id = $1 AND status <> 3 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: PublishContent :execrows
UPDATE contents SET status = 2, published_at = $1, updated_at = $1 WHERE content_id = $2 AND status = 1;

-- name: DeleteContent :execrows
UPDATE contents SET status = 3, updated_at = $1 WHERE content_id = $2 AND status <> 3;

-- name: LockContentForLike :one
SELECT content_id
FROM contents
WHERE content_id = $1
FOR UPDATE;

-- name: AddLike :execrows
INSERT INTO content_likes (content_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RemoveLike :execrows
DELETE FROM content_likes WHERE content_id = $1 AND user_id = $2;

-- name: CountLikes :one
SELECT COUNT(*) FROM content_likes WHERE content_id = $1;

-- name: UpdateLikeCount :exec
UPDATE contents SET like_count = $1, updated_at = $2 WHERE content_id = $3;

-- name: CreateComment :exec
INSERT INTO content_comments (comment_id, content_id, author_user_id, parent_id, text, status)
VALUES ($1, $2, $3, $4, $5, 1);

-- name: IncrementCommentCount :exec
UPDATE contents SET comment_count = comment_count + 1, updated_at = $1 WHERE content_id = $2;

-- name: ListComments :many
SELECT comment_id, content_id, author_user_id, parent_id, text, status, created_at
FROM content_comments WHERE content_id = $1 AND status = 1 ORDER BY created_at ASC LIMIT $2 OFFSET $3;

-- name: DeleteComment :execrows
UPDATE content_comments SET status = 2 WHERE comment_id = $1 AND author_user_id = $2 AND status = 1;

-- name: DecrementCommentCount :exec
UPDATE contents SET comment_count = GREATEST(comment_count - 1, 0), updated_at = $1
WHERE content_id = (SELECT content_id FROM content_comments WHERE comment_id = $2);

-- name: InsertContentOutbox :exec
INSERT INTO content_outbox (event_id, event_type, aggregate_id, topic, partition_key, payload)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ClaimContentOutbox :many
WITH candidates AS (
    SELECT pending.outbox_id
    FROM content_outbox AS pending
    WHERE pending.published_at IS NULL
        AND pending.next_attempt_at <= CURRENT_TIMESTAMP
        AND (pending.claimed_until IS NULL OR pending.claimed_until <= CURRENT_TIMESTAMP)
        AND NOT EXISTS (
            SELECT 1
            FROM content_outbox AS earlier
            WHERE earlier.aggregate_id = pending.aggregate_id
                AND earlier.outbox_id < pending.outbox_id
                AND earlier.published_at IS NULL
        )
    ORDER BY pending.outbox_id
    LIMIT $1
    FOR UPDATE SKIP LOCKED
)
UPDATE content_outbox AS claimed
SET claim_token = $2, claimed_until = $3, attempts = claimed.attempts + 1
FROM candidates
WHERE claimed.outbox_id = candidates.outbox_id
RETURNING claimed.outbox_id, claimed.event_id, claimed.event_type, claimed.aggregate_id,
    claimed.topic, claimed.partition_key, claimed.payload, claimed.attempts;

-- name: MarkContentOutboxPublished :execrows
UPDATE content_outbox
SET published_at = CURRENT_TIMESTAMP, claimed_until = NULL, claim_token = NULL, last_error = ''
WHERE outbox_id = $1 AND claim_token = $2 AND published_at IS NULL;

-- name: RetryContentOutbox :execrows
UPDATE content_outbox
SET next_attempt_at = $3, claimed_until = NULL, claim_token = NULL, last_error = $4
WHERE outbox_id = $1 AND claim_token = $2 AND published_at IS NULL;

-- name: DeletePublishedContentOutbox :execrows
DELETE FROM content_outbox WHERE published_at < $1;
