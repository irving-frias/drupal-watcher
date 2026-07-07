package eventbus

import (
	"sync"
)

type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]func(any)
}

func New() *EventBus {
	return &EventBus{
		handlers: make(map[string][]func(any)),
	}
}

func (b *EventBus) Publish(topic string, event any) {
	b.mu.RLock()
	hs := b.handlers[topic]
	b.mu.RUnlock()
	for _, h := range hs {
		go h(event)
	}
}

func (b *EventBus) Subscribe(topic string, handler func(any)) {
	b.mu.Lock()
	b.handlers[topic] = append(b.handlers[topic], handler)
	b.mu.Unlock()
}

const (
	TopicFileChange = "file.change"
	TopicCacheClear = "cache.clear"
	TopicError      = "error"
)
