package main

import (
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

	address = flag.String("addr", ":8080", "Address for the http server")

	timeout = 1 * time.Second
)

const (
	group   = "example"
	name    = "example"
	version = "1.0"
)

func main() {
	flag.Parse()

	// Create the transport
	trans := mem.New()
	defer trans.Stop()

	// Create the service
	service, _ = mrpc.NewService(
		trans,
		mrpc.WithNGV(name, group, version),
		mrpc.WithStatus(*address), // Enable http status endpoint
	)

	// Request handler
	service.HandleFunc("hi", func(w mrpc.TopicWriter, data []byte) {
		log.Println("Request: hi")
		w.Write([]byte("Hello world"))
	})

	// Start the service
	log.Printf("Starting MRPC service with status on %v", *address)
	ready <- struct{}{}
	log.Fatalf("Service stopped: %v", service.Serve())
}
