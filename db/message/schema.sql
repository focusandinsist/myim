CREATE TABLE conversations (
    conversation_id VARCHAR(36) PRIMARY KEY,
    conversation_type SMALLINT NOT NULL,
    direct_key VARCHAR(73) UNIQUE,
    next_seq BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE conversation_members (
    conversation_id VARCHAR(36) NOT NULL REFERENCES conversations(conversation_id) ON DELETE CASCADE,
    user_id VARCHAR(36) NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (conversation_id, user_id)
);
CREATE TABLE messages (
    message_id VARCHAR(36) PRIMARY KEY,
    request_id VARCHAR(128) NOT NULL,
    conversation_id VARCHAR(36) NOT NULL,
    sender_user_id VARCHAR(36) NOT NULL,
    target_user_id VARCHAR(36) NOT NULL,
    message_type SMALLINT NOT NULL,
    content TEXT NOT NULL,
    sent_at BIGINT NOT NULL,
    seq BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT messages_sender_request_unique UNIQUE (sender_user_id, request_id)
);
