package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"myim/apps/message-service/model"
)

var ErrMessageNotFound = errors.New("message not found") // 按指定条件没有查询到消息

func (d *Dao) SaveMessage(ctx context.Context, message *model.Message) (*model.Message, error) {
	const insertQuery = `
		INSERT INTO messages
			(message_id, request_id, conversation_id, sender_user_id, target_user_id,
			 message_type, content, sent_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (sender_user_id, request_id) DO NOTHING
		RETURNING
			message_id, request_id, conversation_id, sender_user_id, target_user_id,
			message_type, content, sent_at
	`
	persisted := new(model.Message)
	err := d.db.QueryRowContext(
		ctx,
		insertQuery,
		message.MessageID,
		message.RequestID,
		message.ConversationID,
		message.SenderUserID,
		message.TargetUserID,
		message.MessageType,
		message.Content,
		message.SentAt,
	).Scan(
		&persisted.MessageID,
		&persisted.RequestID,
		&persisted.ConversationID,
		&persisted.SenderUserID,
		&persisted.TargetUserID,
		&persisted.MessageType,
		&persisted.Content,
		&persisted.SentAt,
	)
	if err == nil {
		return persisted, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("insert message: %w", err)
	}

	const selectQuery = `
		SELECT
			message_id, request_id, conversation_id, sender_user_id, target_user_id,
			message_type, content, sent_at
		FROM messages
		WHERE sender_user_id = $1 AND request_id = $2
	`
	err = d.db.QueryRowContext(ctx, selectQuery, message.SenderUserID, message.RequestID).Scan(
		&persisted.MessageID,
		&persisted.RequestID,
		&persisted.ConversationID,
		&persisted.SenderUserID,
		&persisted.TargetUserID,
		&persisted.MessageType,
		&persisted.Content,
		&persisted.SentAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMessageNotFound
		}
		return nil, fmt.Errorf("get saved message: %w", err)
	}
	return persisted, nil
}

func (d *Dao) GetMessageByID(ctx context.Context, messageID string) (*model.Message, error) {
	const query = `
		SELECT
			message_id, request_id, conversation_id, sender_user_id, target_user_id,
			message_type, content, sent_at
		FROM messages
		WHERE message_id = $1
	`
	message := new(model.Message)
	if err := d.db.QueryRowContext(ctx, query, messageID).Scan(
		&message.MessageID,
		&message.RequestID,
		&message.ConversationID,
		&message.SenderUserID,
		&message.TargetUserID,
		&message.MessageType,
		&message.Content,
		&message.SentAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMessageNotFound
		}
		return nil, fmt.Errorf("get message: %w", err)
	}
	return message, nil
}

func (d *Dao) DeleteMessage(ctx context.Context, messageID string) error {
	const query = `
		DELETE FROM messages
		WHERE message_id = $1
	`
	result, err := d.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted rows: %w", err)
	}
	if rows == 0 {
		return ErrMessageNotFound
	}
	return nil
}
