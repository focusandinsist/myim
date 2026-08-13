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
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin save message transaction: %w", err)
	}
	defer tx.Rollback()

	const existingQuery = `
		SELECT
			message_id, request_id, conversation_id, sender_user_id, target_user_id,
			message_type, content, sent_at, seq
		FROM messages
		WHERE sender_user_id = $1 AND request_id = $2
	`
	const insertQuery = `
		INSERT INTO messages
			(message_id, request_id, conversation_id, sender_user_id, target_user_id,
			 message_type, content, sent_at, seq)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING
			message_id, request_id, conversation_id, sender_user_id, target_user_id,
			message_type, content, sent_at, seq
	`
	persisted := new(model.Message)
	err = tx.QueryRowContext(ctx, existingQuery, message.SenderUserID, message.RequestID).Scan(
		&persisted.MessageID,
		&persisted.RequestID,
		&persisted.ConversationID,
		&persisted.SenderUserID,
		&persisted.TargetUserID,
		&persisted.MessageType,
		&persisted.Content,
		&persisted.SentAt,
		&persisted.Seq,
	)
	if err == nil {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit existing message transaction: %w", err)
		}
		return persisted, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get existing message: %w", err)
	}

	directKey := message.SenderUserID + ":" + message.TargetUserID
	if message.SenderUserID > message.TargetUserID {
		directKey = message.TargetUserID + ":" + message.SenderUserID
	}
	const conversationQuery = `
		INSERT INTO conversations
			(conversation_id, conversation_type, direct_key)
		VALUES
			($1, $2, $3)
		ON CONFLICT (direct_key) DO UPDATE
		SET updated_at = conversations.updated_at
		RETURNING conversation_id
	`
	if err := tx.QueryRowContext(ctx, conversationQuery, message.ConversationID, model.ConversationTypeDirect, directKey).Scan(&message.ConversationID); err != nil {
		return nil, fmt.Errorf("get or create direct conversation: %w", err)
	}

	const memberQuery = `
		INSERT INTO conversation_members
			(conversation_id, user_id)
		VALUES
			($1, $2),
			($1, $3)
		ON CONFLICT DO NOTHING
	`
	if _, err := tx.ExecContext(ctx, memberQuery, message.ConversationID, message.SenderUserID, message.TargetUserID); err != nil {
		return nil, fmt.Errorf("add direct conversation members: %w", err)
	}
	err = tx.QueryRowContext(ctx, existingQuery, message.SenderUserID, message.RequestID).Scan(
		&persisted.MessageID,
		&persisted.RequestID,
		&persisted.ConversationID,
		&persisted.SenderUserID,
		&persisted.TargetUserID,
		&persisted.MessageType,
		&persisted.Content,
		&persisted.SentAt,
		&persisted.Seq,
	)
	if err == nil {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit concurrent existing message transaction: %w", err)
		}
		return persisted, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get concurrent existing message: %w", err)
	}

	const sequenceQuery = `
		UPDATE conversations
		SET
			next_seq = next_seq + 1,
			updated_at = CURRENT_TIMESTAMP
		WHERE conversation_id = $1
		RETURNING next_seq
	`
	if err := tx.QueryRowContext(ctx, sequenceQuery, message.ConversationID).Scan(&message.Seq); err != nil {
		return nil, fmt.Errorf("allocate message sequence: %w", err)
	}

	err = tx.QueryRowContext(
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
		message.Seq,
	).Scan(
		&persisted.MessageID,
		&persisted.RequestID,
		&persisted.ConversationID,
		&persisted.SenderUserID,
		&persisted.TargetUserID,
		&persisted.MessageType,
		&persisted.Content,
		&persisted.SentAt,
		&persisted.Seq,
	)
	if err != nil {
		return nil, fmt.Errorf("insert message: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit save message transaction: %w", err)
	}
	return persisted, nil
}

func (d *Dao) GetMessageByID(ctx context.Context, messageID string) (*model.Message, error) {
	const query = `
		SELECT
			message_id, request_id, conversation_id, sender_user_id, target_user_id,
			message_type, content, sent_at, seq
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
		&message.Seq,
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
