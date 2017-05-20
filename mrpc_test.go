package mrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestTopicClient(t *testing.T) {
	topic := "test"
	data := "data"
	tr := newFakeTransport()

	c := TopicClient{topic, tr}
	c.Write([]byte("data"))

	d, ok := tr.pubs["test"]
	if !ok {
		t.Fatal("Nothing published to transport")
	}

	if string(d) != data {
		t.Fatalf("Wrong data published to transport: %v", string(d))
	}
}

func TestNewService(t *testing.T) {
	s, err := NewService(&FakeTransport{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if s == nil {
		t.Fatalf("Expected service")
	}
}

func TestNewServiceWithStatus(t *testing.T) {
	trans := newFakeTransport()
	s, err := NewService(trans, EnableStatus("127.0.0.1:8080"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if s == nil {
		t.Fatalf("Expected service")
	}

	if s.statusSrv == nil {
		t.Fatalf("Expected service with status endpoint")
	}

	topic := "pattern"

	// Request to missing topic. Expect error
	if _, err := s.Request(s.GetFQTopic(topic), []byte{}, 1*time.Second); err == nil {
		t.Fatal("Expected error")
	}

	response := "response"
	// Sub to topic
	err = s.HandleFunc(
		topic,
		func(w TopicWriter, d []byte) {
			w.Write([]byte(response))
		},
	)
	if err != nil {
		t.Fatalf("HandleFunc returned unexpected error")
	}
	if _, ok := trans.subs[s.GetFQTopic(topic)]; !ok {
		t.Fatalf("Handler not registered")
	}

	// Req topic
	resp, err := s.Request(s.GetFQTopic(topic), []byte{}, 1*time.Second)
	if err != nil {
		t.Fatalf("Request returned unexpected error: %v", err)
	}
	if string(resp) != response {
		t.Fatalf("Request returned unexpected response: %v", string(resp))
	}

	// Pub to topic
	if err := s.Publish(s.GetFQTopic(topic), []byte{}); err != nil {
		t.Fatalf("Pub to %v: Unexpected error: %v", topic, err)
	}

	// Test http status endpoint
	go s.Serve()

	// Block so the server starts
	time.Sleep(1 * time.Millisecond)

	statusResp, err := http.Get("http://127.0.0.1:8080")
	if err != nil {
		t.Fatalf("HTTP status request returned error: %v", err)
	}
	body, err := ioutil.ReadAll(statusResp.Body)
	if err != nil {
		t.Fatalf("Error reading status response: %v", err)
	}

	status := &statusResponse{}
	err = json.Unmarshal(body, status)
	if err != nil {
		t.Fatalf("HTTP status request returned non json response: %v", err)
	}

	if status.Name != defaultName || status.Version != defaultVer || status.Group != defaultGroup {
		t.Fatalf("Status response not matching: %v", status)
	}

	// Stop the http status server
	s.Stop(nil)
	statusResp, _ = http.Get("http://127.0.0.1:8080")
	if statusResp != nil {
		t.Fatal("Status server not stopped")
	}
}

func TestServeNoStatus(t *testing.T) {
	s, _ := NewService(newFakeTransport())
	if err := s.Serve(); err == nil {
		t.Error("can't serve without status enabled")
	}

	if err := s.Stop(nil); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestNewServiceErr(t *testing.T) {
	_, err := NewService(&FakeTransport{}, func(*Service) error { return fmt.Errorf("Random error") })
	if err == nil {
		t.Fatalf("Expected error: %v", err)
	}
}

func TestGetFQTopic(t *testing.T) {
	testCases := []struct {
		fopt    func(s *Service) error
		topic   string
		fqTopic string
	}{
		{
			func(s *Service) error {
				return nil
			},
			"topic",
			"service.topic",
		},
		{
			func(s *Service) error {
				s.Name = "test"
				return nil
			},
			"topic",
			"test.topic",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			s, _ := NewService(&FakeTransport{}, tc.fopt)
			fqTopic := s.GetFQTopic(tc.topic)

			if tc.fqTopic != fqTopic {
				t.Errorf("Expected %v; Received %v", tc.fqTopic, fqTopic)
			}
		})
	}
}

type FakeTransport struct {
	pubs map[string][]byte
	subs map[string]SubscribeHandlerFunc
}

func newFakeTransport() *FakeTransport {
	return &FakeTransport{
		map[string][]byte{},
		map[string]SubscribeHandlerFunc{},
	}
}

func (t *FakeTransport) Subscribe(topic string, handler SubscribeHandlerFunc) error {
	t.subs[topic] = handler
	return nil
}

func (t *FakeTransport) Publish(topic string, data []byte) error {
	t.pubs[topic] = data
	return nil
}

func (t *FakeTransport) Request(topic string, data []byte, timeout time.Duration) (respData []byte, err error) {
	h, ok := t.subs[topic]
	if !ok {
		return nil, errors.New("So subscribers on this topic")
	}

	r := "replY_topic"
	h(r, topic, data)
	d, ok := t.pubs[r]
	if !ok {
		return nil, errors.New("No response on reply topic")
	}

	return d, nil
}
