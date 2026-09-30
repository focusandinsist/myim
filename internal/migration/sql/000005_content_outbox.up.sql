CREATE TABLE content_outbox (
    outbox_id BIGSERIAL PRIMARY KEY,
    event_id VARCHAR(36) NOT NULL UNIQUE,
    event_type TEXT NOT NULL,
    aggregate_id VARCHAR(36) NOT NULL,
    topic TEXT NOT NULL,
    partition_key TEXT NOT NULL,
    payload BYTEA NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    claimed_until TIMESTAMPTZ,
    claim_token VARCHAR(36),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMPTZ,
    last_error TEXT NOT NULL DEFAULT ''
);
CREATE INDEX content_outbox_pending_idx ON content_outbox (next_attempt_at, outbox_id)
    WHERE published_at IS NULL;
CREATE INDEX content_outbox_aggregate_pending_idx ON content_outbox (aggregate_id, outbox_id)
    WHERE published_at IS NULL;
CREATE INDEX content_outbox_published_idx ON content_outbox (published_at)
    WHERE published_at IS NOT NULL;
