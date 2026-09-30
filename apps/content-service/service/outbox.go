package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"myim/apps/content-service/event"
)

func (s *Service) DispatchOutboxOnce(ctx context.Context) error {
	if s.publisher == nil {
		publisher, err := event.NewSaramaPublisher(s.config.KafkaBrokers)
		if err != nil {
			return fmt.Errorf("connect content event publisher: %w", err)
		}
		s.publisher = publisher
	}
	items, err := s.dao.ClaimOutbox(ctx, s.config.OutboxBatchSize, s.config.OutboxLease)
	if err != nil {
		return err
	}
	var firstError error
	for _, item := range items {
		if err := s.publisher.Publish(item.Topic, item.PartitionKey, item.Payload); err != nil {
			delay := s.config.OutboxRetryBase
			for i := int32(1); i < item.Attempts && delay < 5*time.Minute; i++ {
				delay *= 2
			}
			if delay > 5*time.Minute {
				delay = 5 * time.Minute
			}
			if retryErr := s.dao.RetryOutbox(ctx, item, time.Now().Add(delay), err); retryErr != nil {
				return retryErr
			}
			if firstError == nil {
				firstError = fmt.Errorf("publish content event %s: %w", item.EventID, err)
			}
			continue
		}
		if err := s.dao.MarkOutboxPublished(ctx, item); err != nil {
			return err
		}
	}
	return firstError
}

func (s *Service) RunOutbox(ctx context.Context) {
	ticker := time.NewTicker(s.config.OutboxPollInterval)
	defer ticker.Stop()
	lastCleanup := time.Time{}
	for {
		if ctx.Err() != nil {
			return
		}
		if err := s.DispatchOutboxOnce(ctx); err != nil && ctx.Err() == nil {
			log.Printf("content outbox dispatch: %v", err)
		}
		if time.Since(lastCleanup) >= time.Hour {
			if _, err := s.dao.DeletePublishedOutboxBefore(ctx, time.Now().Add(-s.config.OutboxRetention)); err != nil && ctx.Err() == nil {
				log.Printf("content outbox cleanup: %v", err)
			}
			lastCleanup = time.Now()
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
