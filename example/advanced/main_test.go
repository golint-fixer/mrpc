package main

import (
	"context"
	"os"
	"testing"
)

func TestAdvancedExample(t *testing.T) {
	os.Args = []string{"example"}
	go main()
	<-ready

	// Request to topic a
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	res, err := service.Request(ctx, "service.a", nil)
	if err != nil {
		t.Errorf("Requesting service.a: Unexpected error %v", err)
	}

	if string(res) != "What's up?" {
		t.Errorf("Unexpected response on service.a: %v", string(res))
	}
}
