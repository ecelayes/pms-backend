package redis

import (
	"context"
	"errors"

	"github.com/ecelayes/pms-backend/internal/shared/domain"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// DLQMessage is a type alias for the domain DLQMessage.
type DLQMessage = domain.DLQMessage

// DLQStats is a type alias for the domain DLQStats.
type DLQStats = domain.DLQStats

// DLQStore implements domain.DLQStore using Redis Streams DLQ.
type DLQStore struct {
	client *redis.Client
	logger *zap.Logger
}

var _ domain.DLQStore = (*DLQStore)(nil)

func NewDLQStore(client *redis.Client, logger *zap.Logger) *DLQStore {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &DLQStore{client: client, logger: logger}
}

// GetStats returns the total count of messages currently in the DLQ.
func (s *DLQStore) GetStats(ctx context.Context) (*DLQStats, error) {
	count, err := s.client.XLen(ctx, StreamDLQ).Result()
	if err != nil {
		s.logger.Error("failed to get DLQ stats", zap.Error(err))
		return nil, err
	}
	return &DLQStats{TotalMessages: count}, nil
}

// GetMessages returns up to `limit` messages from the DLQ.
func (s *DLQStore) GetMessages(ctx context.Context, limit int64) (*DLQStats, error) {
	messages, err := s.client.XRange(ctx, StreamDLQ, "-", "+").Result()
	if err != nil {
		s.logger.Error("failed to get DLQ messages", zap.Error(err))
		return nil, err
	}
	dlqMessages := make([]DLQMessage, 0, len(messages))
	for _, msg := range messages {
		if int64(len(dlqMessages)) >= limit {
			break
		}
		dlqMessages = append(dlqMessages, DLQMessage{
			OriginalID: getStringValue(msg.Values, "original_id"),
			EventType:  getStringValue(msg.Values, "event_type"),
			Payload:    getStringValue(msg.Values, "payload"),
			FailedAt:   getInt64Value(msg.Values, "failed_at"),
			Retries:    getIntValue(msg.Values, "retries"),
		})
	}
	return &DLQStats{
		TotalMessages: int64(len(dlqMessages)),
		Messages:      dlqMessages,
	}, nil
}

// DeleteMessage removes a single message by its ID.
func (s *DLQStore) DeleteMessage(ctx context.Context, messageID string) (int64, error) {
	if messageID == "" {
		return 0, errors.New("message ID is required")
	}
	deleted, err := s.client.XDel(ctx, StreamDLQ, messageID).Result()
	if err != nil {
		s.logger.Error("failed to delete DLQ message",
			zap.String("message_id", messageID),
			zap.Error(err))
		return 0, err
	}
	return deleted, nil
}

// Purge removes all messages from the DLQ.
//
// Uses XTrim (MaxLen=0) because go-redis XDel requires explicit IDs and cannot
// accept an empty ID list. XTrim atomically removes all messages, which is what
// a DLQ purge semantically means.
func (s *DLQStore) Purge(ctx context.Context) (int64, error) {
	purged, err := s.client.XTrimMaxLen(ctx, StreamDLQ, 0).Result()
	if err != nil {
		s.logger.Error("failed to purge DLQ", zap.Error(err))
		return 0, err
	}
	s.logger.Info("DLQ purged", zap.Int64("messages_removed", purged))
	return purged, nil
}

func getStringValue(values map[string]interface{}, key string) string {
	if v, ok := values[key].(string); ok {
		return v
	}
	return ""
}

func getInt64Value(values map[string]interface{}, key string) int64 {
	if v, ok := values[key].(int64); ok {
		return v
	}
	if v, ok := values[key].(string); ok {
		var i int64
		for _, c := range v {
			if c >= '0' && c <= '9' {
				i = i*10 + int64(c-'0')
			}
		}
		return i
	}
	return 0
}

func getIntValue(values map[string]interface{}, key string) int {
	if v, ok := values[key].(int); ok {
		return v
	}
	if v, ok := values[key].(int64); ok {
		return int(v)
	}
	if v, ok := values[key].(string); ok {
		var i int
		for _, c := range v {
			if c >= '0' && c <= '9' {
				i = i*10 + int(c-'0')
			}
		}
		return i
	}
	return 0
}
