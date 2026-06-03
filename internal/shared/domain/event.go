package domain

import (
	"sync"
	"time"
)

type DomainEvent interface {
	Name() string
	OccurredOn() time.Time
}
type EventDispatcher struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}
type EventHandler func(event DomainEvent)

var globalDispatcher = &EventDispatcher{
	handlers: make(map[string][]EventHandler),
}

func Subscribe(eventName string, handler EventHandler) {
	globalDispatcher.mu.Lock()
	defer globalDispatcher.mu.Unlock()
	globalDispatcher.handlers[eventName] = append(globalDispatcher.handlers[eventName], handler)
}
func Publish(event DomainEvent) {
	globalDispatcher.mu.RLock()
	defer globalDispatcher.mu.RUnlock()
	handlers, ok := globalDispatcher.handlers[event.Name()]
	if !ok {
		return
	}
	for _, handler := range handlers {
		handler(event)
	}
}

type BaseEvent struct {
	occurredAt time.Time
}

func (e *BaseEvent) OccurredOn() time.Time {
	return e.occurredAt
}
func NewBaseEvent() BaseEvent {
	return BaseEvent{occurredAt: time.Now()}
}
