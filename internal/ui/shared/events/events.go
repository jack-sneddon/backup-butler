// internal/ui/shared/events/events.go
package events

import (
	"sync"

	"github.com/jack-sneddon/backup-butler/internal/ui/shared/viewmodels"
)

type EventType int

const (
	BackupStarted EventType = iota
	BackupProgress
	BackupCompleted
	BackupError
	ConfigChanged
	VersionUpdated
)

type Event struct {
	Type    EventType
	Payload interface{}
}

type Handler func(Event)

type EventBus struct {
	mu       sync.RWMutex
	handlers map[EventType][]Handler
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]Handler),
	}
}

func (b *EventBus) Subscribe(eventType EventType, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *EventBus) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, handler := range b.handlers[event.Type] {
		go handler(event)
	}
}

// Event Publishers
func PublishProgress(bus *EventBus, progress viewmodels.BackupProgress) {
	bus.Publish(Event{
		Type:    BackupProgress,
		Payload: progress,
	})
}

func PublishError(bus *EventBus, err error) {
	bus.Publish(Event{
		Type:    BackupError,
		Payload: err,
	})
}

func PublishBackupComplete(bus *EventBus, result viewmodels.BackupOperation) {
	bus.Publish(Event{
		Type:    BackupCompleted,
		Payload: result,
	})
}
