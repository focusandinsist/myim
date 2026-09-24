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

-- name: AddLike :exec
INSERT INTO content_likes (content_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RemoveLike :exec
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
