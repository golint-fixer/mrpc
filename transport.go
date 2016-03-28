package mrpc

import (
	"time"
)

type SubscribeHandlerFunc func(responseTopic, requestTopic string, rawData []byte)

type Transport interface {
	Connect() error
	Close() error
	Ready() bool

	Subscribe(topicName, channelName string, handler SubscribeHandlerFunc) error
	Publish(topicName string, data []byte) error
	Request(topicName string, data []byte, timeout time.Duration) (respData []byte, err error)
}

func EnsureConnected(t Transport) error {
	if !t.Ready() {
		if err := t.Connect(); err != nil {
			return err
		}
	}
	return nil
}
