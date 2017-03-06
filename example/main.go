package main

import (
	"flag"
	"log"
	"time"

	"github.com/miracl/mrpc"
	"github.com/miracl/mrpc/transport"
)

var (
	address = flag.String("addr", ":8080", "Address for the http server")
	service *mrpc.Service

	timeout = 1 * time.Second
)

const (
	group   = "example"
	name    = "example"
	version = "1.0"
)

func main() {
	flag.Parse()

	// Create the service
	var err error
	service, err = mrpc.NewService(
		transport.NewMem(),
		func(s *mrpc.Service) error {
			s.Group = group
			s.Name = name
			s.Version = version
			return nil
		},
		mrpc.EnableStatus(*address), // Enable http status endpoint
	)
	if err != nil {
		log.Fatal("Service failed to start")
	}

	// Request handler
	service.HandleFunc("hi", func(w mrpc.TopicWriter, data []byte) {
		log.Println("Request: hi")
		w.Write([]byte("Hello world"))
	})

	// Sub handler
	service.HandleFunc("hi.sub", func(w mrpc.TopicWriter, data []byte) {
		log.Printf("Request: hi.sub at %v", string(data))
	})

	// Request handler doing request and pub to the previous handlers
	service.HandleFunc("hi.proxy", func(w mrpc.TopicWriter, data []byte) {
		log.Println("Request: hi.proxy")

		// Example for pub
		err := service.Publish("example.hi.sub", []byte(time.Now().String()))
		if err != nil {
			log.Fatalf("Error publishing to hi.sub: %v", err)
		}

		// Request on another handler
		hw, err := service.Request("example.hi", nil, timeout)
		if err != nil {
			log.Fatalf("Error requesting hi: %v", err)
		}

		w.Write(hw)
	})

	// Start the service
	log.Printf("Starting MRPC service with status on %v", *address)
	log.Fatalf("Service stopped: %v", service.Serve())
}
