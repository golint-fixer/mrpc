package transport

import (
	"fmt"
	"testing"
	"time"
)

func TestMemReq(t *testing.T) {
	msg := "test response"
	trans := NewMem()

	_, err := trans.Request("topic", []byte("test"), 1*time.Second)
	fmt.Println(err)
	if !IsTimeout(err) {
		t.Fatalf("Expected timeout")
	}

	trans.Subscribe("topic", func(resTopic, topic string, data []byte) {
		trans.Publish(resTopic, []byte(msg))
	})

	res, err := trans.Request("topic", []byte("test"), 1*time.Second)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if string(res) != msg {
		t.Fatalf("Unexpected response %v", string(res))
	}
}
