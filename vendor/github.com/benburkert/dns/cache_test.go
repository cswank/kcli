package dns

import (
	"context"
	"math/rand"
	"net"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	client := &Client{
		Resolver: new(Cache),
	}

	localhost := net.IPv4(127, 0, 0, 1).To4()
	ipc := make(chan net.IP, 1)
	ipc <- localhost

	srv := mustServer(HandlerFunc(func(ctx context.Context, w MessageWriter, r *Query) {
		w.Answer("test.local.", time.Minute, &A{A: <-ipc})
	}))

	addrUDP, err := net.ResolveUDPAddr("udp", srv.Addr)
	if err != nil {
		t.Fatal(err)
	}

	query := &Query{
		RemoteAddr: addrUDP,
		Message: &Message{
			Questions: []Question{
				{Name: "test.local.", Type: TypeA},
			},
		},
	}

	msg, err := client.Do(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := localhost, msg.Answers[0].Record.(*A).A.To4(); !want.Equal(got) {
		t.Errorf("want A record %q, got %q", want, got)
	}

	ipc <- net.IPv4(255, 255, 255, 255).To4()

	if msg, err = client.Do(context.Background(), query); err != nil {
		t.Fatal(err)
	}

	if want, got := localhost, msg.Answers[0].Record.(*A).A.To4(); !want.Equal(got) {
		t.Errorf("want A record %q, got %q", want, got)
	}
}

func TestCacheMultiAnswer(t *testing.T) {
	client := &Client{
		Resolver: new(Cache),
	}

	var answered bool
	srv := mustServer(HandlerFunc(func(ctx context.Context, w MessageWriter, r *Query) {
		if answered {
			return
		}

		w.Answer("test.local.", 30*time.Second, &CNAME{CNAME: "cname.test.local."})
		w.Answer("cname.test.local.", time.Minute, &A{A: net.IPv4(127, 0, 1, 1).To4()})
		w.Answer("cname.test.local.", time.Minute, &A{A: net.IPv4(127, 0, 2, 1).To4()})
		w.Answer("cname.test.local.", time.Minute, &A{A: net.IPv4(127, 0, 3, 1).To4()})

		answered = true
	}))

	addrUDP, err := net.ResolveUDPAddr("udp", srv.Addr)
	if err != nil {
		t.Fatal(err)
	}

	query := &Query{
		RemoteAddr: addrUDP,
		Message: &Message{
			Questions: []Question{
				{Name: "test.local.", Type: TypeA},
			},
		},
	}

	msg, err := client.Do(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 4, len(msg.Answers); want != got {
		t.Fatalf("want %d answers, got %d", want, got)
	}
	if want, got := "cname.test.local.", msg.Answers[0].Record.(*CNAME).CNAME; want != got {
		t.Errorf("want CNAME record %q, got %q", want, got)
	}
	if want, got := "127.0.1.1", msg.Answers[1].Record.(*A).A.String(); want != got {
		t.Errorf("want A record %q, got %q", want, got)
	}
	if want, got := "127.0.2.1", msg.Answers[2].Record.(*A).A.String(); want != got {
		t.Errorf("want A record %q, got %q", want, got)
	}
	if want, got := "127.0.3.1", msg.Answers[3].Record.(*A).A.String(); want != got {
		t.Errorf("want A record %q, got %q", want, got)
	}

	rand.Seed(0) // record order: 4 2 3 1

	if msg, err = client.Do(context.Background(), query); err != nil {
		t.Fatal(err)
	}

	if want, got := 4, len(msg.Answers); want != got {
		t.Fatalf("want %d answers, got %d", want, got)
	}
	if want, got := "cname.test.local.", msg.Answers[0].Record.(*CNAME).CNAME; want != got {
		t.Errorf("want CNAME record %q, got %q", want, got)
	}
	if want, got := "127.0.1.1", msg.Answers[1].Record.(*A).A.String(); want != got {
		t.Errorf("want A record %q, got %q", want, got)
	}
	if want, got := "127.0.2.1", msg.Answers[3].Record.(*A).A.String(); want != got {
		t.Errorf("want A record %q, got %q", want, got)
	}
	if want, got := "127.0.3.1", msg.Answers[2].Record.(*A).A.String(); want != got {
		t.Errorf("want A record %q, got %q", want, got)
	}
}
