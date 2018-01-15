package dns

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestPipeline(t *testing.T) {
	stallc := make(chan struct{})
	errc := make(chan error, 1)
	srv := mustServer(HandlerFunc(func(ctx context.Context, w MessageWriter, r *Query) {
		switch r.ID {
		case 1:
			<-stallc

			w.Answer("AAAA.dev.", time.Minute, answers[questions["AAAA"]])
		case 2:
			defer close(stallc)

			w.Answer("A.dev.", time.Minute, answers[questions["A"]])
		default:
			w.Status(NXDomain)

			errc <- errors.New("bad message ID")
		}
	}))

	addr, err := net.ResolveTCPAddr("tcp", srv.Addr)
	if err != nil {
		t.Fatal(err)
	}

	tport := new(Transport)

	conn1, err := tport.DialAddr(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}

	msg := &Message{
		ID:        1,
		Questions: []Question{questions["A"]},
	}

	if err := conn1.Send(msg); err != nil {
		errc <- err
	}

	go func() {
		defer close(errc)

		var msg Message
		if err := conn1.Recv(&msg); err != nil {
			errc <- err
		}
		if msg.ID != 1 {
			errc <- errors.New("message ID mismatch")
		}
	}()

	conn2, err := tport.DialAddr(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}

	msg = &Message{
		ID:        2,
		Questions: []Question{questions["AAAA"]},
	}

	if err := conn2.Send(msg); err != nil {
		errc <- err
	}

	msg = new(Message)
	if err := conn2.Recv(msg); err != nil {
		t.Fatal(err)
	}
	if want, got := 2, msg.ID; want != got {
		t.Errorf("want response message ID %d, got %d", want, got)
	}

	if err := <-errc; err != nil {
		t.Fatal(err)
	}
}
