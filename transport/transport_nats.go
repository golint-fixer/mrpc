package transport

import (
	"time"

	"github.com/miracl/mrpc"
	"github.com/nats-io/nats"
)

type NATSTransport struct {
	conn    *nats.Conn
	servers []string
	init    bool
}

func NewNATSTransport(servers []string) *NATSTransport {
	n := &NATSTransport{
		servers: servers,
		init:    false,
	}

	return n
}

func (n *NATSTransport) Ready() bool {
	return n.init
}

func (n *NATSTransport) Connect() error {
	opts := nats.DefaultOptions
	opts.Servers = n.servers

	var err error

	n.conn, err = opts.Connect()
	if err == nil {
		n.init = true
	}

	return err
}

func (n *NATSTransport) Subscribe(topicName, channelName string, handler mrpc.SubscribeHandlerFunc) error {
	_, err := n.conn.QueueSubscribe(topicName, channelName, func(msg *nats.Msg) {
		handler(msg.Reply, msg.Data)
	})

	return err
}

func (n *NATSTransport) Publish(topicName string, data []byte) error {
	return n.conn.Publish(topicName, data)
}

func (n *NATSTransport) Request(topicName string, data []byte, timeout time.Duration) (respData []byte, err error) {
	msg, err := n.conn.Request(topicName, data, timeout)

	if err != nil {
		if err == nats.ErrTimeout {
			return nil, mrpc.ErrorRequestTimeout
		} else {
			return nil, err
		}
	}

	if msg != nil {
		return msg.Data, nil
	}

	return nil, nil
}

func (n *NATSTransport) Close() error {
	n.conn.Close()

	return nil
}
