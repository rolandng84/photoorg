package sse

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Broker struct {
	mu      sync.RWMutex
	clients map[chan Event]struct{}
}

func NewBroker() *Broker {
	return &Broker{
		clients: make(map[chan Event]struct{}),
	}
}

func (b *Broker) Subscribe() chan Event {
	ch := make(chan Event, 256)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	log.Debug().Int("clients", len(b.clients)).Msg("SSE client subscribed")
	return ch
}

func (b *Broker) Unsubscribe(ch chan Event) {
	b.mu.Lock()
	delete(b.clients, ch)
	close(ch)
	b.mu.Unlock()
	log.Debug().Int("clients", len(b.clients)).Msg("SSE client unsubscribed")
}

func (b *Broker) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- event:
		default:
			log.Warn().Str("type", event.Type).Msg("SSE client too slow, dropped event")
		}
	}
}

func (b *Broker) PublishJSON(eventType string, payload interface{}) {
	b.Publish(Event{Type: eventType, Payload: payload})
}
