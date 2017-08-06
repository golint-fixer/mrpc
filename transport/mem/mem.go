package mem

import (
	"context"
	"math/rand"
	"strconv"
	"sync"

	"github.com/miracl/mrpc"
)

type subscribeHandlerFunc func(responseTopic, requestTopic string, data []byte)

type msg struct {
	topic string
	data  []byte
}

// Mem is a transport layer that is entirely in the processes memory.
// Mem is intended for testing.
type Mem struct {
	mu   sync.RWMutex
	subs map[string]chan msg
}

// New creates new in memory transport
func New() *Mem {
	return &Mem{
		mu:   sync.RWMutex{},
		subs: map[string]chan msg{},
	}
}

// Subscribe add a handler for topic
func (t *Mem) Subscribe(topic string, handler mrpc.SubscribeHandlerFunc) error {
	ch := t.getOrCreateChan(topic)
	go func() {
		for m := range ch {
			go handler(m.topic, topic, m.data)
		}
	}()
	return nil
}

// Publish checks if there is a subscriber and executes it
func (t *Mem) Publish(topic string, data []byte) error {
	ch, ok := t.getChan(topic)
	if ok {
		ch <- msg{data: data}
	}
	return nil
}

// Request check if there is a subscriber, executes it and returns the result
func (t *Mem) Request(ctx context.Context, topic string, data []byte) ([]byte, error) {
	ch, ok := t.getChan(topic)
	if !ok {
		return nil, context.DeadlineExceeded
	}

	resTopic := "_RESPONSE." + topic + "." + strconv.Itoa(rand.Int())
	resCh := t.getOrCreateChan(resTopic)
	ch <- msg{topic: resTopic, data: data}

	select {
	case d := <-resCh:
		return d.data, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Stop stops all subscribers and reset the transport
func (t *Mem) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, ch := range t.subs {
		close(ch)
	}
	t.subs = map[string]chan msg{}
}

func (t *Mem) getChan(topic string) (chan<- msg, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	ch, ok := t.subs[topic]
	return ch, ok
}

func (t *Mem) getOrCreateChan(topic string) <-chan msg {
	t.mu.Lock()
	ch, ok := t.subs[topic]
	if !ok {
		ch = make(chan msg)
		t.subs[topic] = ch
	}
	t.mu.Unlock()
	return ch
}
