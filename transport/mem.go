package transport

import (
	"sync"
	"time"

	"github.com/miracl/mrpc"
)

type subscribeHandlerFunc func(responseTopic, requestTopic string, data []byte)

// Mem is a transport layer that is entirely in the processes memory.
// Mem is intended for testing.
type Mem struct {
	subs map[string]mrpc.SubscribeHandlerFunc
	mu   sync.Mutex
}

// NewMem creates new in memory transport
func NewMem() *Mem {
	subs := map[string]mrpc.SubscribeHandlerFunc{}
	return &Mem{subs, sync.Mutex{}}
}

// Subscribe add a handler for topic
func (t *Mem) Subscribe(topic string, handler mrpc.SubscribeHandlerFunc) error {
	t.mu.Lock()
	t.subs[topic] = handler
	t.mu.Unlock()
	return nil
}

// Publish checks if there is a subscriber and executes it
func (t *Mem) Publish(topic string, data []byte) error {
	t.mu.Lock()
	h, ok := t.subs[topic]
	t.mu.Unlock()
	if ok {
		h("", topic, data)
	}
	return nil
}

// Request check if there is a subscriber, executes it and returns the result
func (t *Mem) Request(topic string, data []byte, timeout time.Duration) ([]byte, error) {
	resTopic := "_RESPONSE." + topic

	resCh := make(chan []byte, 1)
	t.Subscribe(
		resTopic,
		func(resTopic, topic string, data []byte) {
			resCh <- data
		},
	)

	go func() {
		t.mu.Lock()
		h, ok := t.subs[topic]
		t.mu.Unlock()
		if ok {
			h(resTopic, topic, data)
		}
	}()

	timeoutCh := make(chan bool, 1)
	go func() {
		time.Sleep(timeout)
		timeoutCh <- true
	}()

	select {
	case data = <-resCh:
		return data, nil
	case <-timeoutCh:
		return nil, errTimeout
	}
}
