package main

import (
	"encoding/json"
	"flag"
	"log"
	"time"

	"github.com/miracl/mrpc"
	"github.com/miracl/mrpc/transport/mem"
)

var (
	// Service is global so it can be used in the test.
	service *mrpc.Service
	ready   = make(chan struct{}, 1)

	timeout = 1 * time.Second
)

type dataWithMeta struct {
	Data          []byte
	ResponseTopic string
}

// Example applications where A makes a request which is going trough couple of
// services (B, C, D) and the last one in the chain responds directly to A.
//
// A   B   C   D
// #-->|   |   |
// #   #   |   |
// #   |-->|   |
// #   |   #   |
// #   |   |-->|
// #   |   |   #
// #<--|---|---|
// |   |   |   |
func main() {
	flag.Parse()

	// Create the service
	service, _ = mrpc.NewService(mem.New())
	ready <- struct{}{}

	// Request handler
	service.HandleFunc("a", func(w mrpc.TopicWriter, data []byte) {
		logRequest("a")

		// Request blocking topic B
		response, _ := service.Request(service.GetFQTopic("b"), data, timeout)
		w.Write(response)

		logDone("a")
	})

	service.HandleFunc("b", func(w mrpc.TopicWriter, data []byte) {
		logRequest("b")

		// Add the response topic to the request, so D will respond to B
		respData, _ := json.Marshal(dataWithMeta{data, w.Topic()})
		service.Publish(service.GetFQTopic("c"), respData)

		logDone("b")
	})

	service.HandleFunc("c", func(w mrpc.TopicWriter, data []byte) {
		logRequest("c")

		service.Publish(service.GetFQTopic("d"), data)

		logDone("c")
	})

	service.HandleFunc("d", func(w mrpc.TopicWriter, data []byte) {
		logRequest("d")

		v := &dataWithMeta{}
		json.Unmarshal(data, v)

		service.Publish(v.ResponseTopic, []byte("What's up?"))

		logDone("d")
	})

	// Start the service
	log.Printf("Starting MRPC service")
	log.Fatalf("Service stopped: %v", service.Serve())
}

func logRequest(topic string) {
	log.Printf("[REQ] %v", topic)
}

func logDone(topic string) {
	log.Printf("[DONE] %v", topic)
}
