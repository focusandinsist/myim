-- name: CreateUser :exec
INSERT INTO users (user_id, user_name, password, phone, email, nickname, avatar, bio,
    gender, birthday, region, status, register_time, last_login_time, updated_time)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);

-- name: GetUserByUserID :one
SELECT user_id, user_name, password, phone, email, nickname, avatar, bio,
    gender, birthday, region, status, register_time, last_login_time, updated_time
FROM users WHERE user_id = $1;

-- name: GetUserByUserName :one
SELECT user_id, user_name, password, phone, email, nickname, avatar, bio,
    gender, birthday, region, status, register_time, last_login_time, updated_time
FROM users WHERE LOWER(user_name) = LOWER($1);

-- name: UserNameExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(user_name) = LOWER($1));

-- name: UpdateLastLoginTime :execrows
UPDATE users SET last_login_time = $1 WHERE user_id = $2;
