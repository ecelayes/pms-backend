package domain

import (
	"sync"
	"testing"
	"time"
)

func TestBaseEventOccurredOn(t *testing.T) {
	now := time.Now()
	event := BaseEvent{occurredAt: now}
	
	if !event.OccurredOn().Equal(now) {
		t.Errorf("Expected %v, got %v", now, event.OccurredOn())
	}
}

func TestNewBaseEvent(t *testing.T) {
	before := time.Now()
	event := NewBaseEvent()
	after := time.Now()
	
	if event.occurredAt.Before(before) || event.occurredAt.After(after) {
		t.Error("OccurredAt should be between before and after calls")
	}
}

func TestEventDispatcherSubscribeAndPublish(t *testing.T) {
	dispatcher := &EventDispatcher{
		handlers: make(map[string][]EventHandler),
	}
	
	testEvent := &testDomainEvent{name: "test.event"}
	handled := false
	var mu sync.Mutex
	
	dispatcher.mu.Lock()
	dispatcher.handlers["test.event"] = append(dispatcher.handlers["test.event"], func(e DomainEvent) {
		mu.Lock()
		handled = true
		mu.Unlock()
	})
	dispatcher.mu.Unlock()
	
	// Simulate publish
	dispatcher.mu.RLock()
	handlers := dispatcher.handlers["test.event"]
	dispatcher.mu.RUnlock()
	
	if len(handlers) != 1 {
		t.Errorf("Expected 1 handler, got %d", len(handlers))
	}
	
	for _, h := range handlers {
		h(testEvent)
	}
	
	mu.Lock()
	if !handled {
		t.Error("Expected handler to be called")
	}
	mu.Unlock()
}

func TestEventDispatcherNoHandlers(t *testing.T) {
	dispatcher := &EventDispatcher{
		handlers: make(map[string][]EventHandler),
	}
	
	testEvent := &testDomainEvent{name: "nonexistent.event"}
	handled := false
	
	dispatcher.mu.RLock()
	handlers, ok := dispatcher.handlers[testEvent.Name()]
	dispatcher.mu.RUnlock()
	
	if ok || len(handlers) > 0 {
		t.Error("Should have no handlers for nonexistent event")
	}
	
	for _, h := range handlers {
		h(testEvent)
		handled = true
	}
	
	if handled {
		t.Error("Handler should not be called for nonexistent event")
	}
}

type testDomainEvent struct {
	name string
}

func (e *testDomainEvent) Name() string {
	return e.name
}

func (e *testDomainEvent) OccurredOn() time.Time {
	return time.Now()
}

func TestGlobalDispatcher(t *testing.T) {
	// Save original global
	orig := globalDispatcher
	defer func() { globalDispatcher = orig }()
	
	// Replace with clean dispatcher
	globalDispatcher = &EventDispatcher{
		handlers: make(map[string][]EventHandler),
	}
	
	receivedEvent := false
	var mu sync.Mutex
	
	Subscribe("global.test", func(e DomainEvent) {
		mu.Lock()
		receivedEvent = true
		mu.Unlock()
	})
	
	Publish(&testDomainEvent{name: "global.test"})
	
	mu.Lock()
	if !receivedEvent {
		t.Error("Expected global event to be received")
	}
	mu.Unlock()
}

func TestGlobalDispatcherNoHandler(t *testing.T) {
	// Save original global
	orig := globalDispatcher
	defer func() { globalDispatcher = orig }()
	
	// Replace with clean dispatcher
	globalDispatcher = &EventDispatcher{
		handlers: make(map[string][]EventHandler),
	}
	
	// Should not panic when publishing to unknown event
	Publish(&testDomainEvent{name: "unknown.event"})
}

func TestSubscribeMultipleHandlers(t *testing.T) {
	// Save original global
	orig := globalDispatcher
	defer func() { globalDispatcher = orig }()
	
	// Replace with clean dispatcher
	globalDispatcher = &EventDispatcher{
		handlers: make(map[string][]EventHandler),
	}
	
	count := 0
	var mu sync.Mutex
	
	Subscribe("multi.test", func(e DomainEvent) {
		mu.Lock()
		count++
		mu.Unlock()
	})
	Subscribe("multi.test", func(e DomainEvent) {
		mu.Lock()
		count++
		mu.Unlock()
	})
	
	Publish(&testDomainEvent{name: "multi.test"})
	
	mu.Lock()
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
	mu.Unlock()
}
