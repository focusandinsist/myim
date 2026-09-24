-- name: GetMessageByRequest :one
SELECT message_id, request_id, conversation_id, sender_user_id, target_user_id,
    message_type, content, sent_at, seq
FROM messages WHERE sender_user_id = $1 AND request_id = $2;

-- name: GetMessageByID :one
SELECT message_id, request_id, conversation_id, sender_user_id, target_user_id,
    message_type, content, sent_at, seq
FROM messages WHERE message_id = $1;

-- name: GetOrCreateConversation :one
INSERT INTO conversations (conversation_id, conversation_type, direct_key)
VALUES ($1, $2, $3)
ON CONFLICT (direct_key) DO UPDATE SET updated_at = conversations.updated_at
RETURNING conversation_id;

-- name: AddConversationMembers :exec
INSERT INTO conversation_members (conversation_id, user_id)
VALUES ($1, $2), ($1, $3) ON CONFLICT DO NOTHING;

-- name: AllocateMessageSequence :one
UPDATE conversations SET next_seq = next_seq + 1, updated_at = CURRENT_TIMESTAMP
WHERE conversation_id = $1 RETURNING next_seq;

-- name: InsertMessage :one
INSERT INTO messages (message_id, request_id, conversation_id, sender_user_id, target_user_id,
    message_type, content, sent_at, seq)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING message_id, request_id, conversation_id, sender_user_id, target_user_id,
    message_type, content, sent_at, seq;

-- name: DeleteMessage :execrows
DELETE FROM messages WHERE message_id = $1;
