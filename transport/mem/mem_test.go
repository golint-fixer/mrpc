package mem

import (
	"context"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestPubSub(t *testing.T) {
	msg := "test message "
	trans := New()
	defer trans.Stop()

	res := make(chan []byte)
	err := trans.Subscribe("topic", func(resTopic, topic string, data []byte) {
		res <- data
		close(res)
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := trans.Publish("topic", []byte(msg)); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	d := <-res
	if string(d) != msg {
		t.Fatalf("Unexpected response: %v", string(d))
	}
}
func TestMemReq(t *testing.T) {
	msg := "test response"
	trans := New()
	defer trans.Stop()

	trans.Subscribe("topic", func(resTopic, topic string, data []byte) {
		trans.Publish(resTopic, []byte(msg))
	})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			res, err := trans.Request(ctx, "topic", []byte("test"))
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if string(res) != msg {
				t.Errorf("Unexpected response %v", string(res))
			}
		}()
	}

	wg.Wait()
}

func TestMemReqTimeout(t *testing.T) {
	trans := New()
	defer trans.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := trans.Request(ctx, "topic", []byte("test"))
	if err != context.DeadlineExceeded {
		t.Fatalf("Expected timeout")
	}

	trans.Subscribe("sleep", func(resTopic, topic string, data []byte) {
		time.Sleep(150 * time.Millisecond)
		trans.Publish(resTopic, nil)
	})

	ctx2, cancel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel2()
	_, err = trans.Request(ctx2, "sleep", []byte("test"))
	if err != context.DeadlineExceeded {
		t.Fatalf("Expected timeout")
	}
}

func TestMemClose(t *testing.T) {
	trans := New()

	done := make(chan struct{})
	trans.Subscribe("sleep", func(resTopic, topic string, data []byte) {
		time.Sleep(1 * time.Millisecond)
		trans.Publish(resTopic, nil)
		close(done)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	_, err := trans.Request(ctx, "sleep", nil)
	if err != context.DeadlineExceeded {
		t.Fatalf("Expected timeout")
	}
	trans.Stop()

	<-done
}

func BenchmarkMemReqOneSub(b *testing.B) {
	trans := New()
	defer trans.Stop()

	trans.Subscribe("topic", func(resTopic, topic string, data []byte) {
		trans.Publish(resTopic, nil)
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, cancel := context.WithTimeout(context.Background(), 0)
			defer cancel()
			trans.Request(ctx, "topic", nil)
		}
	})
}

func BenchmarkMemReqManySubs(b *testing.B) {
	cases := []struct {
		name string
		subs int
	}{
		{
			name: "1 subscriber",
			subs: 1,
		},
		{
			name: "10 subscribers",
			subs: 10,
		},
		{
			name: "100 subscribers",
			subs: 100,
		},
		{
			name: "1k subscribers",
			subs: 10000,
		},
		{
			name: "1M subscribers",
			subs: 100000,
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			trans := New()
			defer trans.Stop()

			for i := 0; i < tc.subs; i++ {
				trans.Subscribe("topic"+strconv.Itoa(i), func(resTopic, topic string, data []byte) {
					trans.Publish(resTopic, nil)
				})
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					t := "topic" + strconv.Itoa(rand.Intn(tc.subs))
					ctx, cancel := context.WithTimeout(context.Background(), 0)
					defer cancel()
					trans.Request(ctx, t, nil)
				}
			})
		})
	}
}
