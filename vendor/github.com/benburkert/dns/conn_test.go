package dns

import (
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func TestPacketConnRoundTrip(t *testing.T) {
	tests := []struct {
		name string

		req *Message
		res *Message
		err error
	}{
		{
			name: "happy-path",

			req: &Message{
				Questions: []Question{
					{
						Name:  "example.com.",
						Type:  TypeA,
						Class: ClassIN,
					},
				},
			},
			res: &Message{
				Answers: []Resource{
					{
						Name:  "example.com.",
						Class: ClassIN,
						TTL:   60 * time.Second,
						Record: &A{
							A: net.IPv4(127, 0, 0, 1).To4(),
						},
					},
				},
				Questions: []Question{
					{
						Name:  "example.com.",
						Type:  TypeA,
						Class: ClassIN,
					},
				},
			},
		},
		{
			name: "oversized-query",

			req: &Message{
				Questions: []Question{
					{
						Name:  strings.Repeat(strings.Repeat("a", 63)+".", 10),
						Type:  TypeA,
						Class: ClassIN,
					},
				},
			},
			err: ErrOversizedMessage,
		},
	}

	t.Parallel()

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			c1, c2 := net.Pipe()

			client := &PacketConn{
				Conn: c1,
			}
			server := &PacketConn{
				Conn: c2,
			}

			err := testRoundTrip(client, server, test.req, test.res)
			if want, got := test.err, err; want != got {
				if want != nil {
					t.Fatalf("want error %q, got %q", want, got)
				}
				t.Error(err)
			}
		})
	}
}

func TestStreamConnRoundTrip(t *testing.T) {
	tests := []struct {
		name string

		req *Message
		res *Message
		err error
	}{
		{
			name: "happy-path",

			req: &Message{
				Questions: []Question{
					{
						Name:  "example.com.",
						Type:  TypeA,
						Class: ClassIN,
					},
				},
			},
			res: &Message{
				Answers: []Resource{
					{
						Name:  "example.com.",
						Class: ClassIN,
						TTL:   60 * time.Second,
						Record: &A{
							A: net.IPv4(127, 0, 0, 1).To4(),
						},
					},
				},
				Questions: []Question{
					{
						Name:  "example.com.",
						Type:  TypeA,
						Class: ClassIN,
					},
				},
			},
		},
	}

	t.Parallel()

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			c1, c2 := net.Pipe()

			client := &StreamConn{
				Conn: c1,
			}
			server := &StreamConn{
				Conn: c2,
			}

			err := testRoundTrip(client, server, test.req, test.res)
			if want, got := test.err, err; want != got {
				if want != nil {
					t.Fatalf("want error %q, got %q", want, got)
				}
				t.Error(err)
			}
		})
	}
}

func testRoundTrip(client, server Conn, req, res *Message) error {
	var (
		g errgroup.Group
	)

	g.Go(func() error {
		defer client.Close()

		if err := client.Send(req); err != nil {
			return err
		}

		msg := new(Message)
		if err := client.Recv(msg); err != nil {
			return err
		}

		if want, got := res, msg; !reflect.DeepEqual(want, got) {
			return fmt.Errorf("want response message %+v, got %+v", want, got)
		}

		return nil
	})

	g.Go(func() error {
		defer server.Close()

		msg := new(Message)
		if err := server.Recv(msg); err != nil {
			return err
		}

		if want, got := req, msg; !reflect.DeepEqual(want, got) {
			return fmt.Errorf("want request message %+v, got %+v", want, got)
		}

		return server.Send(res)
	})

	return g.Wait()
}
