package mrpc_test

import (
	"context"
	"testing"
	"time"

	"github.com/miracl/mrpc"
	memTrans "github.com/miracl/mrpc/transport/mem"
	natsTrans "github.com/miracl/mrpc/transport/nats"
	nats "github.com/nats-io/go-nats"
)

func TestConcurrentRequests(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatal(err)
	}

	for name, tr := range map[string]mrpc.Transport{
		"mem":  memTrans.New(),
		"nats": natsTrans.New(nc),
	} {
		t.Run(name, func(t *testing.T) {
			service, err := mrpc.NewService(tr)
			if err != nil {
				t.Fatal(err)
			}

			service.HandleFunc("sync", func(w mrpc.TopicWriter, data []byte) {
				w.Write([]byte("done"))
			})

			service.HandleFunc("slow", func(w mrpc.TopicWriter, data []byte) {
				time.Sleep(1 * time.Millisecond)
				w.Write([]byte("done"))
			})

			go service.Serve()
			// Wait for the service to start
			if _, err := service.Request(context.Background(), "service.sync", nil); err != nil {
				t.Fatal(err)
			}

			if err := service.Publish("service.slow", []byte{}); err != nil {
				t.Fatal(err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1999*time.Microsecond)
			defer cancel()
			if _, err := service.Request(ctx, "service.slow", nil); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func BenchmarkConcurrency(b *testing.B) {

	cases := []struct {
		name      string
		transport string
	}{
		{
			name:      "memory",
			transport: "memory",
		},
		{
			name:      "nats",
			transport: "nats",
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {

			var trans mrpc.Transport
			switch tc.transport {
			case "memory":
				trans = memTrans.New()
			case "nats":
				nc, err := nats.Connect(nats.DefaultURL)
				if err != nil {
					b.Fatal(err)
				}
				trans = natsTrans.New(nc)
			}

			service, err := mrpc.NewService(trans)
			if err != nil {
				b.Fatal(err)
			}

			service.HandleFunc("sync", func(w mrpc.TopicWriter, data []byte) {
				w.Write([]byte("done"))
			})

			service.HandleFunc("topic", func(w mrpc.TopicWriter, data []byte) {
				time.Sleep(100 * time.Millisecond)
				w.Write([]byte("done"))
			})

			go service.Serve()
			// Wait for the service to start
			if _, err := service.Request(context.Background(), "service.sync", nil); err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if _, err := service.Request(context.Background(), "service.topic", nil); err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}
