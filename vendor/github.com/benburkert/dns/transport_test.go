package dns

import (
	"context"
	"crypto/tls"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/benburkert/dns/internal/must"
)

var transportTests = []struct {
	name string

	req *Message

	res *Message
}{
	{
		name: "single-A-match",

		req: &Message{
			ID:        1,
			Questions: []Question{questions["A"]},
		},

		res: &Message{
			ID:        1,
			Response:  true,
			Questions: []Question{questions["A"]},
			Answers: []Resource{
				{
					Name:   "A.dev.",
					Class:  ClassIN,
					TTL:    60 * time.Second,
					Record: answers[questions["A"]],
				},
			},
		},
	},
	{
		name: "single-AAAA-match",

		req: &Message{
			ID:        2,
			Questions: []Question{questions["AAAA"]},
		},

		res: &Message{
			ID:        2,
			Response:  true,
			Questions: []Question{questions["AAAA"]},
			Answers: []Resource{
				{
					Name:   "AAAA.dev.",
					Class:  ClassIN,
					TTL:    60 * time.Second,
					Record: answers[questions["AAAA"]],
				},
			},
		},
	},
}

func TestTransport(t *testing.T) {
	t.Parallel()

	srv := mustServer(&answerHandler{answers})

	t.Run("udp", func(t *testing.T) {
		t.Parallel()

		addr, err := net.ResolveUDPAddr("udp", srv.Addr)
		if err != nil {
			t.Fatal(err)
		}

		testTransport(t, new(Transport), addr)
	})

	t.Run("tcp", func(t *testing.T) {
		t.Parallel()

		addr, err := net.ResolveTCPAddr("tcp", srv.Addr)
		if err != nil {
			t.Fatal(err)
		}

		testTransport(t, new(Transport), addr)
	})

	t.Run("tcp-tls", func(t *testing.T) {
		t.Parallel()

		ca := must.CACert("ca.dev", nil)

		srv.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{
				*must.LeafCert("dns-server.dev", ca).TLS(),
				*ca.TLS(),
			},
		}

		ln, err := net.Listen("tcp", ":0")
		if err != nil {
			t.Fatal(err)
		}

		go srv.ServeTLS(context.Background(), ln)

		tport := &Transport{
			TLSConfig: &tls.Config{
				ServerName: "dns-server.dev",
				RootCAs:    must.CertPool(ca.TLS()),
			},
		}

		testTransport(t, tport, OverTLSAddr{ln.Addr()})
	})
}

func TestTransportProxy(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}

	tport := &Transport{
		Proxy: func(_ context.Context, _ net.Addr) (net.Addr, error) {
			return ln.Addr(), nil
		},
	}

	conn, err := tport.DialAddr(context.Background(), new(net.TCPAddr))
	if err != nil {
		t.Fatal(err)
	}

	if want, got := ln.Addr().(*net.TCPAddr).Port, conn.RemoteAddr().(*net.TCPAddr).Port; want != got {
		t.Errorf("want dialed addr %q, got %q", want, got)
	}
}

func testTransport(t *testing.T, tport *Transport, addr net.Addr) {
	for _, test := range transportTests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			conn, err := tport.DialAddr(context.Background(), addr)
			if err != nil {
				t.Fatal(err)
			}

			if err := conn.Send(test.req); err != nil {
				t.Fatal(err)
			}

			var msg Message
			if err := conn.Recv(&msg); err != nil {
				t.Fatal(err)
			}

			if want, got := test.res, &msg; !reflect.DeepEqual(want, got) {
				t.Errorf("want response %+v, got %+v", want, got)
			}
		})
	}
}

var (
	questions = map[string]Question{
		"A": {
			Name:  "A.dev.",
			Type:  TypeA,
			Class: ClassIN,
		},
		"AAAA": {
			Name:  "AAAA.dev.",
			Type:  TypeAAAA,
			Class: ClassIN,
		},
	}

	answers = map[Question]Record{
		questions["A"]: &A{
			A: net.IPv4(127, 0, 0, 1).To4(),
		},
		questions["AAAA"]: &AAAA{
			AAAA: net.ParseIP("::1"),
		},
	}
)

type answerHandler struct {
	Answers map[Question]Record
}

func (a *answerHandler) ServeDNS(ctx context.Context, w MessageWriter, r *Query) {
	for _, q := range r.Questions {
		if answer, ok := a.Answers[q]; ok {
			w.Answer(q.Name, time.Minute, answer)
		}
	}
}
