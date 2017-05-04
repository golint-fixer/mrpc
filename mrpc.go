package mrpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	defaultGroup = ""
	defaultName  = "service"
	defaultVer   = "0.1"
)

// TopicHandler serves MRPC requests
type TopicHandler interface {
	Serve(w TopicWriter, requestTopic string, data []byte)
}

// TopicWriter writes data to topic
type TopicWriter interface {
	Write(data []byte) error
	Topic() string
}

// The SubscribeHandlerFunc type is an adapter to allow the use of ordinary
// functions as MRPC handlers.
type SubscribeHandlerFunc func(responseTopic, requestTopic string, data []byte)

// Transport is a way of making MPRC requests
type Transport interface {
	Subscribe(topic string, handler SubscribeHandlerFunc) error
	Publish(topic string, data []byte) error
	Request(topic string, data []byte, timeout time.Duration) (respData []byte, err error)
}

// TopicClient is implementation of TopicWriter
type TopicClient struct {
	topic     string
	transport Transport
}

// Write writes data to the transport
func (t *TopicClient) Write(data []byte) error {
	return t.transport.Publish(t.topic, data)
}

// Topic returns the clients topic as a string
func (t *TopicClient) Topic() string {
	return t.topic
}

// Service is MRPC topic router
type Service struct {
	Group   string
	Name    string
	Version string

	T Transport

	Adapter MessageAdapter

	statusServer *http.Server
	quitChannel  chan os.Signal
}

// NewService returns new MRPC service
func NewService(t Transport, opts ...func(*Service) error) (*Service, error) {
	s := Service{
		Group:   defaultGroup,
		Name:    defaultName,
		Version: defaultVer,

		T: t,

		Adapter: &emptyMessageAdapter{},

		quitChannel: make(chan os.Signal),
	}

	for _, opt := range opts {
		if err := opt(&s); err != nil {
			return nil, err
		}
	}

	return &s, nil
}

// GetFQTopic returns the full topic name
func (s *Service) GetFQTopic(topic string) string {
	return fmt.Sprintf("%s.%s", s.Name, topic)
}

// Handle registers a handler for particular topic
func (s *Service) Handle(topic string, handler TopicHandler) error {
	return s.T.Subscribe(
		s.GetFQTopic(topic),
		func(responseTopic, requestTopic string, data []byte) {
			s.Adapter.ProcessMessage(subMessageType, requestTopic, data)
			handler.Serve(&TopicClient{responseTopic, s.T}, requestTopic, data)
		},
	)
}

// HandleFunc registers a handler functions for particular topic
func (s *Service) HandleFunc(pattern string, handler func(TopicWriter, []byte)) error {
	return s.Handle(pattern, HandlerFunc(handler))
}

// HandlerFunc is function implementing the TopicHandler interface
type HandlerFunc func(TopicWriter, []byte)

// Serve calls the HandlerFunc
func (f HandlerFunc) Serve(w TopicWriter, requestTopic string, data []byte) {
	f(w, data)
}

// Publish to a topic
func (s *Service) Publish(topic string, data []byte) (err error) {
	s.Adapter.ProcessMessage(pubMessageType, topic, data)
	return s.T.Publish(topic, data)
}

// Request does a MRPC request and waits for response
func (s *Service) Request(topic string, data []byte, timeout time.Duration) (respData []byte, err error) {
	s.Adapter.ProcessMessage(reqMessageType, topic, data)
	respData, err = s.T.Request(topic, data, timeout)
	if err != nil {
		return nil, err
	}

	s.Adapter.ProcessMessage(resMessageType, topic, respData)
	return respData, err
}

// Serve start the MRPC server
func (s *Service) Serve() error {
	// Start status http service if enabled
	if s.statusServer != nil {
		go s.statusServer.ListenAndServe()
	}

	signal.Notify(s.quitChannel, os.Interrupt)
	sig := <-s.quitChannel
	return fmt.Errorf("Signal received: %v", sig)
}

// EnableStatus returns can be added to NewService call to enable status
func EnableStatus(addr string) func(*Service) error {
	return func(s *Service) error {
		s.statusServer = &http.Server{Addr: addr, Handler: s}
		return nil
	}
}

type statusResponse struct {
	Name    string `json:"name"`
	Group   string `json:"group"`
	Version string `json:"version"`
}

// ServeHTTP serves status information on http
func (s *Service) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	enc.Encode(statusResponse{s.Name, s.Group, s.Version})
	w.Header().Set("Content-Type", "application/json")
}
