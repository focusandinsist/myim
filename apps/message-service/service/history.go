package service

import (
	"context"
	"errors"
	"net/http"

	proto_message "myim/api/protobuf/message"
	"myim/apps/message-service/dao"
	"myim/internal/auth"

	"github.com/google/uuid"
)

const defaultHistoryPageSize int32 = 20
const maxHistoryPageSize int32 = 100

func (s *Service) HandleCGConversationList(ctx context.Context, input *proto_message.CGConversationList, token string) (*proto_message.GCConversationList, error) {
	output := new(proto_message.GCConversationList)
	claims, err := auth.Validate(token, s.config.Auth.Secret)
	if err != nil {
		output.ErrorCode = http.StatusUnauthorized
		return output, err
	}
	if input == nil || (input.GetAfterConversationId() != "" && uuid.Validate(input.GetAfterConversationId()) != nil) || input.GetPageSize() < 0 || input.GetPageSize() > maxHistoryPageSize {
		output.ErrorCode = http.StatusBadRequest
		return output, errors.New("invalid conversation cursor or page_size")
	}
	pageSize := input.GetPageSize()
	if pageSize == 0 {
		pageSize = defaultHistoryPageSize
	}
	rows, err := s.dao.ListUserConversations(ctx, claims.UserID, input.GetAfterConversationId(), pageSize+1)
	if err != nil {
		output.ErrorCode = http.StatusInternalServerError
		return output, err
	}
	output.HasMore = len(rows) > int(pageSize)
	if output.HasMore {
		rows = rows[:int(pageSize)]
	}
	for _, row := range rows {
		output.Conversations = append(output.Conversations, &proto_message.Conversation{ConversationId: row.ConversationID, PeerUserId: row.PeerUserID, LatestSeq: row.LatestSeq})
	}
	if len(rows) > 0 {
		output.NextConversationId = rows[len(rows)-1].ConversationID
	} else {
		output.NextConversationId = input.GetAfterConversationId()
	}
	output.ErrorMsg = "ok"
	return output, nil
}

func (s *Service) HandleCGMessageHistory(ctx context.Context, input *proto_message.CGMessageHistory, token string) (*proto_message.GCMessageHistory, error) {
	output := new(proto_message.GCMessageHistory)
	claims, err := auth.Validate(token, s.config.Auth.Secret)
	if err != nil {
		output.ErrorCode = http.StatusUnauthorized
		return output, err
	}
	if input == nil || uuid.Validate(input.GetConversationId()) != nil || input.GetAfterSeq() < 0 || input.GetPageSize() < 0 || input.GetPageSize() > maxHistoryPageSize {
		output.ErrorCode = http.StatusBadRequest
		return output, errors.New("invalid conversation_id, after_seq or page_size")
	}
	pageSize := input.GetPageSize()
	if pageSize == 0 {
		pageSize = defaultHistoryPageSize
	}
	rows, err := s.dao.ListConversationMessages(ctx, claims.UserID, input.GetConversationId(), input.GetAfterSeq(), pageSize+1)
	if errors.Is(err, dao.ErrConversationNotFound) {
		output.ErrorCode = http.StatusNotFound
		return output, err
	}
	if err != nil {
		output.ErrorCode = http.StatusInternalServerError
		return output, err
	}
	output.HasMore = len(rows) > int(pageSize)
	if output.HasMore {
		rows = rows[:int(pageSize)]
	}
	output.NextSeq = input.GetAfterSeq()
	for _, row := range rows {
		output.Messages = append(output.Messages, &proto_message.Message{
			MessageId: row.MessageID, RequestId: row.RequestID, ConversationId: row.ConversationID,
			SenderUserId: row.SenderUserID, TargetUserId: row.TargetUserID,
			MessageType: proto_message.MessageType(row.MessageType), Content: row.Content,
			SentAt: row.SentAt, Seq: row.Seq,
		})
		output.NextSeq = row.Seq
	}
	output.ErrorMsg = "ok"
	return output, nil
}
