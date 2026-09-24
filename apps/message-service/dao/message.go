package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"myim/apps/message-service/model"
	messagedb "myim/internal/db/message"
)

var ErrMessageNotFound = errors.New("message not found")

func (d *Dao) SaveMessage(ctx context.Context, message *model.Message) (*model.Message, error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin save message transaction: %w", err)
	}
	defer tx.Rollback()
	queries := d.queries.WithTx(tx)
	row, err := queries.GetMessageByRequest(ctx, messagedb.GetMessageByRequestParams{SenderUserID: message.SenderUserID, RequestID: message.RequestID})
	if err == nil {
		if err = tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit existing message transaction: %w", err)
		}
		return messageFromRow(row), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get existing message: %w", err)
	}
	directKey := message.SenderUserID + ":" + message.TargetUserID
	if message.SenderUserID > message.TargetUserID {
		directKey = message.TargetUserID + ":" + message.SenderUserID
	}
	conversationID, err := queries.GetOrCreateConversation(ctx, messagedb.GetOrCreateConversationParams{ConversationID: message.ConversationID, ConversationType: model.ConversationTypeDirect, DirectKey: sql.NullString{String: directKey, Valid: true}})
	if err != nil {
		return nil, fmt.Errorf("get or create direct conversation: %w", err)
	}
	message.ConversationID = conversationID
	if err = queries.AddConversationMembers(ctx, messagedb.AddConversationMembersParams{ConversationID: conversationID, UserID: message.SenderUserID, UserID_2: message.TargetUserID}); err != nil {
		return nil, fmt.Errorf("add direct conversation members: %w", err)
	}
	if row, err = queries.GetMessageByRequest(ctx, messagedb.GetMessageByRequestParams{SenderUserID: message.SenderUserID, RequestID: message.RequestID}); err == nil {
		if err = tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit concurrent existing message transaction: %w", err)
		}
		return messageFromRow(row), nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get concurrent existing message: %w", err)
	}
	message.Seq, err = queries.AllocateMessageSequence(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("allocate message sequence: %w", err)
	}
	inserted, err := queries.InsertMessage(ctx, messagedb.InsertMessageParams{MessageID: message.MessageID, RequestID: message.RequestID, ConversationID: message.ConversationID, SenderUserID: message.SenderUserID, TargetUserID: message.TargetUserID, MessageType: int16(message.MessageType), Content: message.Content, SentAt: message.SentAt, Seq: message.Seq})
	if err != nil {
		return nil, fmt.Errorf("insert message: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit save message transaction: %w", err)
	}
	return &model.Message{MessageID: inserted.MessageID, RequestID: inserted.RequestID, ConversationID: inserted.ConversationID, SenderUserID: inserted.SenderUserID, TargetUserID: inserted.TargetUserID, MessageType: int32(inserted.MessageType), Content: inserted.Content, SentAt: inserted.SentAt, Seq: inserted.Seq}, nil
}

func (d *Dao) GetMessageByID(ctx context.Context, messageID string) (*model.Message, error) {
	row, err := d.queries.GetMessageByID(ctx, messageID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrMessageNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get message: %w", err)
	}
	return &model.Message{MessageID: row.MessageID, RequestID: row.RequestID, ConversationID: row.ConversationID, SenderUserID: row.SenderUserID, TargetUserID: row.TargetUserID, MessageType: int32(row.MessageType), Content: row.Content, SentAt: row.SentAt, Seq: row.Seq}, nil
}

func (d *Dao) DeleteMessage(ctx context.Context, messageID string) error {
	rows, err := d.queries.DeleteMessage(ctx, messageID)
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
	}
	if rows == 0 {
		return ErrMessageNotFound
	}
	return nil
}

func messageFromRow(row messagedb.GetMessageByRequestRow) *model.Message {
	return &model.Message{MessageID: row.MessageID, RequestID: row.RequestID, ConversationID: row.ConversationID, SenderUserID: row.SenderUserID, TargetUserID: row.TargetUserID, MessageType: int32(row.MessageType), Content: row.Content, SentAt: row.SentAt, Seq: row.Seq}
}
