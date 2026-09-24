-- name: UserExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1 AND status = 0);

-- name: SetFollow :one
INSERT INTO follows (follower_user_id, followee_user_id, status)
VALUES ($1, $2, $3)
ON CONFLICT (follower_user_id, followee_user_id) DO UPDATE SET
    status = EXCLUDED.status, updated_at = CURRENT_TIMESTAMP
RETURNING follower_user_id, followee_user_id, status, created_at, updated_at;

-- name: ListFollowing :many
SELECT follower_user_id, followee_user_id, status, created_at, updated_at FROM follows
WHERE follower_user_id = $1 AND status = 1
ORDER BY updated_at DESC, followee_user_id ASC LIMIT $2 OFFSET $3;

-- name: ListFollowers :many
SELECT follower_user_id, followee_user_id, status, created_at, updated_at FROM follows
WHERE followee_user_id = $1 AND status = 1
ORDER BY updated_at DESC, follower_user_id ASC LIMIT $2 OFFSET $3;

-- name: IsFollowing :one
SELECT EXISTS(SELECT 1 FROM follows WHERE follower_user_id = $1 AND followee_user_id = $2 AND status = 1);
