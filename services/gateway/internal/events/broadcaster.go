package events

import (
	"sync"

	"github.com/satellite-communications/gateway/internal/model"
)

type Broadcaster struct {
	mu          sync.RWMutex
	subscribers map[chan model.GatewayEvent]struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[chan model.GatewayEvent]struct{}),
	}
}

func (b *Broadcaster) Subscribe() chan model.GatewayEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan model.GatewayEvent, 100)
	b.subscribers[ch] = struct{}{}

	return ch
}

func (b *Broadcaster) Unsubscribe(ch chan model.GatewayEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.subscribers[ch]; exists {
		delete(b.subscribers, ch)
		close(ch)
	}
}

func (b *Broadcaster) Publish(event model.GatewayEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		select {
		case ch <- event:
		default:
			// Subscriber's buffer is full.
			// Drop this event rather than block the gateway.
		}
	}
}
