package mem

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"sync"

	"github.com/miracl/mrpc/transport"
)

func TestMemReq(t *testing.T) {
	msg := "test response"
	trans := New()

	_, err := trans.Request("topic", []byte("test"), 1*time.Second)
	if !transport.IsTimeout(err) {
		t.Fatalf("Expected timeout")
	}

	trans.Subscribe("sleep", func(resTopic, topic string, data []byte) {
		time.Sleep(150 * time.Millisecond)
		trans.Publish(resTopic, []byte(msg))
	})
	_, err = trans.Request("sleep", []byte("test"), 100*time.Millisecond)
	if !transport.IsTimeout(err) {
		t.Fatalf("Expected timeout")
	}

	trans.Subscribe("topic", func(resTopic, topic string, data []byte) {
		trans.Publish(resTopic, []byte(msg))
	})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := trans.Request("topic", []byte("test"), 1*time.Second)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if string(res) != msg {
				t.Fatalf("Unexpected response %v", string(res))
			}
		}()
	}

	wg.Wait()
	trans.Stop()
}

func BenchmarkMemReq(b *testing.B) {
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
					trans.Request(t, nil, 1*time.Second)
				}
			})
		})
	}
}
