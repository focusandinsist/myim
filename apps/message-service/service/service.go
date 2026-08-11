package service

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	proto_message "myim/api/protobuf/message"
	"myim/apps/message-service/config"
	"myim/apps/message-service/dao"
	"myim/apps/message-service/model"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type client struct {
	userID   string          // 已通过JWT认证的用户ID
	conn     *websocket.Conn // WebSocket连接
	send     chan []byte     // 单写协程消费的发送队列
	done     chan struct{}   // 连接停止信号
	stopOnce sync.Once       // 保证连接只关闭一次
	mu       sync.Mutex      // 保护关闭状态和发送队列
	stopped  bool            // 连接是否已经停止
}

type Service struct {
	config  *config.Config     // message service配置
	dao     *dao.Dao           // message PostgreSQL数据访问对象
	clients map[string]*client // 当前实例中的用户连接
	mu      sync.RWMutex       // 保护用户连接集合
	stopped bool               // 服务是否已经停止接收连接
}

func New(c *config.Config, dao *dao.Dao) *Service {
	return &Service{
		config:  c,
		dao:     dao,
		clients: make(map[string]*client),
	}
}

func (s *Service) HandleCGConnection(ctx context.Context, userID string, conn *websocket.Conn) {
	client := &client{
		userID: userID,
		conn:   conn,
		send:   make(chan []byte, s.config.SendQueueSize),
		done:   make(chan struct{}),
	}

	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		client.stop()
		return
	}
	previous := s.clients[userID]
	s.clients[userID] = client
	s.mu.Unlock()
	if previous != nil {
		previous.stop()
	}

	go s.writeMessages(client)
	s.readMessages(ctx, client)

	s.mu.Lock()
	if s.clients[userID] == client {
		delete(s.clients, userID)
	}
	s.mu.Unlock()
	client.stop()
}

func (s *Service) readMessages(ctx context.Context, client *client) {
	client.conn.SetReadLimit(s.config.ReadLimit)
	client.conn.SetReadDeadline(time.Now().Add(s.config.PongWait))
	client.conn.SetPongHandler(func(string) error {
		return client.conn.SetReadDeadline(time.Now().Add(s.config.PongWait))
	})

	for {
		messageType, payload, err := client.conn.ReadMessage()
		if err != nil {
			return
		}
		if messageType != websocket.BinaryMessage {
			s.sendError(client, "", http.StatusUnsupportedMediaType, "protobuf binary message is required")
			continue
		}

		input := new(proto_message.CGWSMessage)
		if err := proto.Unmarshal(payload, input); err != nil {
			s.sendError(client, "", http.StatusBadRequest, "invalid protobuf message")
			continue
		}
		if input.GetAction() != proto_message.WSAction_WS_ACTION_SEND_MESSAGE {
			s.sendError(client, input.GetRequestId(), http.StatusBadRequest, "unsupported websocket action")
			continue
		}
		if strings.TrimSpace(input.GetRequestId()) == "" || len(input.GetRequestId()) > 128 {
			s.sendError(client, "", http.StatusBadRequest, "request_id must contain 1 to 128 bytes")
			continue
		}
		if uuid.Validate(input.GetTargetUserId()) != nil {
			s.sendError(client, input.GetRequestId(), http.StatusBadRequest, "valid target_user_id is required")
			continue
		}
		content := input.GetContent()
		if strings.TrimSpace(content) == "" || utf8.RuneCountInString(content) > 4096 {
			s.sendError(client, input.GetRequestId(), http.StatusBadRequest, "content must contain 1 to 4096 characters")
			continue
		}
		if input.GetMessageType() != proto_message.MessageType_MESSAGE_TYPE_TEXT {
			s.sendError(client, input.GetRequestId(), http.StatusBadRequest, "only text messages are supported")
			continue
		}

		conversationID := client.userID + ":" + input.GetTargetUserId()
		if client.userID > input.GetTargetUserId() {
			conversationID = input.GetTargetUserId() + ":" + client.userID
		}
		persisted, err := s.dao.SaveMessage(ctx, &model.Message{
			MessageID:      uuid.NewString(),
			RequestID:      input.GetRequestId(),
			ConversationID: conversationID,
			SenderUserID:   client.userID,
			TargetUserID:   input.GetTargetUserId(),
			MessageType:    int32(input.GetMessageType()),
			Content:        content,
			SentAt:         time.Now().UnixMilli(),
		})
		if err != nil {
			s.sendError(client, input.GetRequestId(), http.StatusInternalServerError, "persist message failed")
			continue
		}
		if persisted.TargetUserID != input.GetTargetUserId() ||
			persisted.MessageType != int32(input.GetMessageType()) ||
			persisted.Content != content {
			s.sendError(client, input.GetRequestId(), http.StatusConflict, "request_id has already been used")
			continue
		}

		message := &proto_message.Message{
			MessageId:    persisted.MessageID,
			RequestId:    persisted.RequestID,
			SenderUserId: persisted.SenderUserID,
			TargetUserId: persisted.TargetUserID,
			MessageType:  proto_message.MessageType(persisted.MessageType),
			Content:      persisted.Content,
			SentAt:       persisted.SentAt,
		}

		s.mu.RLock()
		target := s.clients[message.TargetUserId]
		s.mu.RUnlock()
		if target == nil {
			client.enqueue(&proto_message.GCWSMessage{
				ErrorMsg:  "ok",
				Action:    proto_message.WSAction_WS_ACTION_MESSAGE_ACK,
				RequestId: input.GetRequestId(),
				Message:   message,
			})
			continue
		}
		if !target.enqueue(&proto_message.GCWSMessage{
			ErrorMsg:  "ok",
			Action:    proto_message.WSAction_WS_ACTION_PUSH_MESSAGE,
			RequestId: input.GetRequestId(),
			Message:   message,
		}) {
			client.enqueue(&proto_message.GCWSMessage{
				ErrorMsg:  "ok",
				Action:    proto_message.WSAction_WS_ACTION_MESSAGE_ACK,
				RequestId: input.GetRequestId(),
				Message:   message,
			})
			continue
		}
		client.enqueue(&proto_message.GCWSMessage{
			ErrorMsg:  "ok",
			Action:    proto_message.WSAction_WS_ACTION_MESSAGE_ACK,
			RequestId: input.GetRequestId(),
			Message:   message,
		})
	}
}

func (s *Service) writeMessages(client *client) {
	ticker := time.NewTicker(s.config.PingPeriod)
	defer ticker.Stop()

	for {
		select {
		case payload := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(s.config.WriteWait))
			if err := client.conn.WriteMessage(websocket.BinaryMessage, payload); err != nil {
				client.stop()
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(s.config.WriteWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				client.stop()
				return
			}
		case <-client.done:
			return
		}
	}
}

func (s *Service) sendError(client *client, requestID string, errorCode int32, errorMessage string) {
	client.enqueue(&proto_message.GCWSMessage{
		ErrorCode: errorCode,
		ErrorMsg:  errorMessage,
		Action:    proto_message.WSAction_WS_ACTION_ERROR,
		RequestId: requestID,
	})
}

func (c *client) enqueue(output *proto_message.GCWSMessage) bool {
	payload, err := proto.Marshal(output)
	if err != nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stopped {
		return false
	}
	select {
	case c.send <- payload:
		return true
	default:
		return false
	}
}

func (c *client) stop() {
	c.stopOnce.Do(func() {
		c.mu.Lock()
		c.stopped = true
		close(c.done)
		c.mu.Unlock()
		c.conn.Close()
	})
}

func (s *Service) Stop() {
	s.mu.Lock()
	s.stopped = true
	clients := make([]*client, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	s.clients = make(map[string]*client)
	s.mu.Unlock()
	for _, client := range clients {
		client.stop()
	}
}
