package broker

import (
	"fmt"
	"sync"
)

type SimpleBroker struct {
	mu          sync.Mutex
	subscribers []chan any
}

func New() *SimpleBroker {
	return &SimpleBroker{
		mu:          sync.Mutex{},
		subscribers: make([]chan any, 0),
	}
}

func (b *SimpleBroker) Subscribe() (chan any, error) {
	ch := make(chan any)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers = append(b.subscribers, ch)
	return ch, nil
}

func (b *SimpleBroker) Unsubscribe(ch chan any) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, c := range b.subscribers {
		if c == ch {
			b.subscribers = append(b.subscribers[:i], b.subscribers[i+1:]...)
			close(c)
			return nil
		}
	}
	return fmt.Errorf("channel not found")
}

func (b *SimpleBroker) Publish(m any) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subscribers {
		ch <- m
	}
	return nil
}
