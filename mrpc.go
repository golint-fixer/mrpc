// Service running as handlers on message queue topics

package mrpc

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	CHAN_NAME = "process"

	DefaultRequestTimeout    = 1 * time.Second
	DefaultErrorHandlerTopic = "errors"
)

var (
	ErrorRequestTimeout = errors.New("mrpc: Request Timeout")
)

// Objects implementing the TopicHandler interface can be
// registered to serve a particular topic.
type TopicHandler interface {
	Serve(w TopicWriter, requestTopic string, data []byte)
}

// A TopicWriter interface is used for create response for bidirectional
// handlers.
type TopicWriter interface {
	Write(data []byte) error
}

// Implementation of TopicWriter
type TopicClient struct {
	topic     string
	transport Transport
}

func (t *TopicClient) Write(data []byte) error {
	return t.transport.Publish(t.topic, data)
}

type ServiceOptions struct {
	RequestTimeout    time.Duration
	ErrorHandlerTopic string
}

var DefaultOptions = ServiceOptions{
	RequestTimeout:    DefaultRequestTimeout,
	ErrorHandlerTopic: DefaultErrorHandlerTopic,
}

type Service struct {
	quitChannel chan os.Signal

	transport Transport
	Opts      ServiceOptions

	serviceGroup   string
	serviceName    string
	serviceVersion string
}

func NewService(transport Transport, serviceGroup, serviceName, serviceVersion string, options *ServiceOptions) *Service {
	quitChan := make(chan os.Signal)

	var opts ServiceOptions
	if options == nil {
		opts = DefaultOptions
	} else {
		opts = *options
	}

	return &Service{
		quitChannel:    quitChan,
		transport:      transport,
		Opts:           opts,
		serviceGroup:   serviceGroup,
		serviceName:    serviceName,
		serviceVersion: serviceVersion,
	}
}

func (s *Service) GetFQTopic(topic string) string {
	return fmt.Sprintf("%s.%s", s.serviceName, topic)
}

func (s *Service) Handle(topic string, handler TopicHandler) error {
	if err := EnsureConnected(s.transport); err != nil {
		return err
	}

	s.transport.Subscribe(s.GetFQTopic(topic), CHAN_NAME, func(responseTopic, requestTopic string, data []byte) {
		handler.Serve(&TopicClient{responseTopic, s.transport}, requestTopic, data)
	})

	return nil
}

func (s *Service) HandleFunc(pattern string, handler func(TopicWriter, []byte)) {
	s.Handle(pattern, HandlerFunc(handler))
}

func (s *Service) Serve() os.Signal {
	signal.Notify(s.quitChannel, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGABRT)
	sig := <-s.quitChannel
	return sig
}

type HandlerFunc func(TopicWriter, []byte)

func (f HandlerFunc) Serve(w TopicWriter, requestTopic string, data []byte) {
	f(w, data)
}
