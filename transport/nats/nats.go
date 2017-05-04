package nats

import (
	"time"

	"github.com/nats-io/go-nats"

	"github.com/miracl/mrpc"
	"github.com/miracl/mrpc/transport"
)

const queue = "mrpc"

// natsConn is interface implemented by nats.Conn. It is in use to enable testing
type natsConn interface {
	QueueSubscribe(subj, queue string, cb nats.MsgHandler) (*nats.Subscription, error)
	Publish(subj string, data []byte) error
	Request(subj string, data []byte, timeout time.Duration) (m *nats.Msg, err error)
	Close()
}

// NATS is implementation for the mrpc transport with NATS bus
type NATS struct {
	Conn  natsConn
	Queue string
}

// New is constructor for NATS
func New(conn natsConn) *NATS {
	return &NATS{Conn: conn, Queue: queue}
}

// Subscribe adds a handler to particular topic
func (n *NATS) Subscribe(topicName string, handler mrpc.SubscribeHandlerFunc) error {
	_, err := n.Conn.QueueSubscribe(topicName, n.Queue, func(msg *nats.Msg) {
		handler(msg.Reply, topicName, msg.Data)
	})
	return err
}

// Publish publishes the data to the given topic
func (n *NATS) Publish(topicName string, data []byte) error {
	return n.Conn.Publish(topicName, data)
}

// Request send a request to a topic and waits for response
func (n *NATS) Request(topicName string, data []byte, timeout time.Duration) (respData []byte, err error) {
	msg, err := n.Conn.Request(topicName, data, timeout)
	if err != nil {
		if err == nats.ErrTimeout {
			return nil, transport.ErrTimeout
		}
		return nil, err
	}

	if msg == nil {
		return nil, nil
	}
	return msg.Data, nil
}

// Close closes nats connection
func (n *NATS) Close() {
	n.Conn.Close()
}
