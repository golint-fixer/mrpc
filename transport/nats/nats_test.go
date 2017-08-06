package nats

import (
	"context"
	"fmt"
	"testing"
	"time"

	nats "github.com/nats-io/go-nats"
)

func TestNewNats(t *testing.T) {
	testCases := []struct {
		conn natsConn
	}{
		{
			&nats.Conn{},
		},
		{
			&natsConnMock{},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			trans := New(tc.conn)
			if trans.Conn == nil {
				t.Fatalf("Transport doesn't have nats connection")
			}
		})
	}
}

func TestPubSub(t *testing.T) {
	trans := NATS{Conn: newNatsConnMock()}

	trans.Subscribe("test", func(responseTopic, requestTopic string, rawData []byte) {
		fmt.Println(responseTopic, requestTopic)
		trans.Publish(responseTopic, []byte("Sub:Response"))
	})

	connMock := trans.Conn.(*natsConnMock)
	_, ok := connMock.subs["test"]
	if !ok {
		t.Fatalf("QueueSubscribe not called")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	resData, err := trans.Request(ctx, "test", []byte{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if string(resData) != "Sub:Response" {
		t.Fatalf("Unexpected request response")
	}

	trans.Close()
	if !trans.Conn.(*natsConnMock).closed {
		t.Fatalf("Transport not closed")
	}
}

func TestNatsRequest(t *testing.T) {
	trans := NATS{Conn: newNatsConnMock()}

	// Test nats Timeout
	trans.Conn.(*natsConnMock).reqerr = context.DeadlineExceeded
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	_, err := trans.Request(ctx, "test", []byte{})
	fmt.Println(err)
	if err != context.DeadlineExceeded {
		t.Fatalf("Expected timeout: %v", err)
	}
	trans.Conn.(*natsConnMock).reqerr = nil

	// Test nil message from nats
	trans.Conn.(*natsConnMock).reqmsgnil = true
	ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel2()
	msg, err := trans.Request(ctx2, "test", []byte{})
	if msg != nil {
		t.Fatalf("Expected nil message: %v", err)
	}
	trans.Conn.(*natsConnMock).reqmsgnil = false

	// Test unknown error
	ctx3, cancel3 := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel3()
	_, err = trans.Request(ctx3, "test", []byte{})
	if err == nil {
		t.Fatalf("Expected error")
	}
}

type natsConnMock struct {
	subs map[string]nats.MsgHandler
	pubs map[string][]byte

	reqerr    error
	reqmsgnil bool

	closed bool
}

func newNatsConnMock() *natsConnMock {
	c := &natsConnMock{}
	c.subs = map[string]nats.MsgHandler{}
	c.pubs = map[string][]byte{}
	return c
}

func (conn *natsConnMock) QueueSubscribe(subj, queue string, cb nats.MsgHandler) (*nats.Subscription, error) {
	conn.subs[subj] = cb
	return nil, nil
}

func (conn *natsConnMock) Publish(subj string, data []byte) error {
	conn.pubs[subj] = data
	return nil
}

func (conn *natsConnMock) RequestWithContext(ctx context.Context, subj string, data []byte) (m *nats.Msg, err error) {
	if conn.reqerr != nil {
		return nil, conn.reqerr
	}

	if conn.reqmsgnil {
		return nil, nil
	}

	replyTopic := "reply"

	h, ok := conn.subs[subj]
	if !ok {
		return nil, fmt.Errorf("No subscribers")
	}

	h(&nats.Msg{Reply: replyTopic})

	redData, ok := conn.pubs[replyTopic]
	if !ok {
		return nil, context.DeadlineExceeded
	}
	return &nats.Msg{Data: redData}, nil
}

func (conn *natsConnMock) Close() {
	conn.closed = true
}
