package domain

import (
	"context"
	"time"
)

// EventPublisher publishes domain events to a messaging backend (Redis pub/sub, Kafka, etc).
//
// Implementations live in adapter packages. The application/domain layers depend on this
// interface, never on the concrete Redis type.
type EventPublisher interface {
	PublishReservationCreated(ctx context.Context, payload ReservationCreatedPayload) error
	PublishReservationConfirmed(ctx context.Context, payload ReservationConfirmedPayload) error
	PublishReservationCancelled(ctx context.Context, payload ReservationCancelledPayload) error
}

// EventSubscriber subscribes to domain events.
//
// EventSubscriber is intentionally simple: each implementation owns its own goroutine
// lifecycle. Pass a context to control cancellation.
type EventSubscriber interface {
	SubscribeReservationCreated(ctx context.Context, handler func(payload ReservationCreatedPayload))
	SubscribeReservationCancelled(ctx context.Context, handler func(payload ReservationCancelledPayload))
}

// StreamProducer publishes messages to a durable stream (Redis Streams, Kafka, etc).
//
// Streams provide at-least-once delivery with consumer groups, ack/nack, and DLQ.
type StreamProducer interface {
	PublishReservationCreated(ctx context.Context, payload ReservationCreatedPayload) error
	PublishReservationConfirmed(ctx context.Context, payload ReservationConfirmedPayload) error
	PublishReservationCancelled(ctx context.Context, payload ReservationCancelledPayload) error
	EnsureGroups(ctx context.Context) error
}

// StreamEventHandler processes a single message from a stream. Return nil to ack,
// return an error to nack/retry.
type StreamEventHandler func(msg *StreamMessage) error

// StreamConsumer reads messages from a stream and dispatches to handlers.
//
// Consume blocks until ctx is cancelled.
type StreamConsumer interface {
	Consume(ctx context.Context, eventTypes []string, handler StreamEventHandler)
}

// AnalyticsRecorder stores domain events for later aggregation/analysis.
type AnalyticsRecorder interface {
	RecordEvent(ctx context.Context, eventType string, payload interface{}) error
	GetEventCounts(ctx context.Context, days int) (map[string]int64, error)
}

// DLQStore provides access to the dead-letter queue (failed messages).
type DLQStore interface {
	GetStats(ctx context.Context) (*DLQStats, error)
	GetMessages(ctx context.Context, limit int64) (*DLQStats, error)
	DeleteMessage(ctx context.Context, messageID string) (int64, error)
	Purge(ctx context.Context) (int64, error)
}


// StreamMessage is the message envelope used by stream consumers.
type StreamMessage struct {
	ID        string
	EventType string
	Payload   []byte
	Retries   int
}

// DLQStats holds summary statistics for the dead-letter queue.
type DLQStats struct {
	TotalMessages int64       `json:"total_messages"`
	Messages      []DLQMessage `json:"messages,omitempty"`
}

// DLQMessage represents a single message in the dead-letter queue.
type DLQMessage struct {
	OriginalID string `json:"original_id"`
	EventType  string `json:"event_type"`
	Payload    string `json:"payload"`
	FailedAt   int64  `json:"failed_at"`
	Retries    int    `json:"retries"`
}


// AnalyticsEvent represents a single event stored in the analytics stream.
type AnalyticsEvent struct {
	ID         string                 `json:"id"`
	EventType  string                 `json:"event_type"`
	PropertyID string                 `json:"property_id"`
	UnitTypeID string                 `json:"unit_type_id"`
	GuestEmail string                 `json:"guest_email"`
	Timestamp  time.Time              `json:"timestamp"`
	Data       map[string]interface{} `json:"data,omitempty"`
}
