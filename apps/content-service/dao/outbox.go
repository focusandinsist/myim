package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"myim/apps/content-service/model"
	contentdb "myim/db/content"

	"github.com/google/uuid"
)

func (d *Dao) ClaimOutbox(ctx context.Context, limit int32, lease time.Duration) ([]model.OutboxEvent, error) {
	claimToken := uuid.NewString()
	rows, err := d.queries.ClaimContentOutbox(ctx, contentdb.ClaimContentOutboxParams{
		Limit: limit, ClaimToken: sql.NullString{String: claimToken, Valid: true},
		ClaimedUntil: sql.NullTime{Time: time.Now().Add(lease), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("claim content outbox: %w", err)
	}
	result := make([]model.OutboxEvent, 0, len(rows))
	for _, row := range rows {
		result = append(result, model.OutboxEvent{
			OutboxID: row.OutboxID, EventID: row.EventID, EventType: row.EventType,
			AggregateID: row.AggregateID, Topic: row.Topic, PartitionKey: row.PartitionKey,
			Payload: row.Payload, Attempts: row.Attempts, ClaimToken: claimToken,
		})
	}
	return result, nil
}

func (d *Dao) MarkOutboxPublished(ctx context.Context, item model.OutboxEvent) error {
	rows, err := d.queries.MarkContentOutboxPublished(ctx, contentdb.MarkContentOutboxPublishedParams{
		OutboxID: item.OutboxID, ClaimToken: sql.NullString{String: item.ClaimToken, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("mark content outbox published: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("content outbox claim expired for event %s", item.EventID)
	}
	return nil
}

func (d *Dao) RetryOutbox(ctx context.Context, item model.OutboxEvent, next time.Time, publishErr error) error {
	message := publishErr.Error()
	if len(message) > 1024 {
		message = message[:1024]
	}
	rows, err := d.queries.RetryContentOutbox(ctx, contentdb.RetryContentOutboxParams{
		OutboxID: item.OutboxID, ClaimToken: sql.NullString{String: item.ClaimToken, Valid: true},
		NextAttemptAt: next, LastError: message,
	})
	if err != nil {
		return fmt.Errorf("retry content outbox: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("content outbox claim expired for event %s", item.EventID)
	}
	return nil
}

func (d *Dao) DeletePublishedOutboxBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	rows, err := d.queries.DeletePublishedContentOutbox(ctx, sql.NullTime{Time: cutoff, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("delete published content outbox: %w", err)
	}
	return rows, nil
}
