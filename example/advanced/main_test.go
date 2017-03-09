package main

import (
	"os"
	"testing"
)

func TestAdvancedExample(t *testing.T) {
	os.Args = []string{"example"}
	go main()
	<-ready

	// Request to topic a
	res, err := service.Request("service.a", nil, timeout)
	if err != nil {
		t.Errorf("Requesting service.a: Unexpected error %v", err)
	}

	if string(res) != "What's up?" {
		t.Errorf("Unexpected response on service.a: %v", string(res))
	}
}
