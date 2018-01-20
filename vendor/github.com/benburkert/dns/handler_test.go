package dns

import (
	"context"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestResolveMux(t *testing.T) {
	t.Parallel()

	mailZone := &Zone{
		Origin: "mx.",
		TTL:    24 * time.Hour,
		RRs: map[string][]Record{
			"foo": {
				&MX{
					Pref: 101,
					MX:   "a.foo.mx.",
				},
				&MX{
					Pref: 101,
					MX:   "b.foo.mx.",
				},
			},
			"bar": {
				&MX{
					Pref: 101,
					MX:   "a.bar.mx.",
				},
				&MX{
					Pref: 101,
					MX:   "b.bar.mx.",
				},
			},
		},
	}

	mux := new(ResolveMux)
	mux.Handle(TypeMX, ".", mailZone)
	mux.Handle(TypeANY, "localhost.", localhostZone)

	client := &Client{
		Resolver: mux,
	}

	srv := mustServer(HandlerFunc(Refuse))
	addr, err := net.ResolveUDPAddr("udp", srv.Addr)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("single MX question query", func(t *testing.T) {
		t.Parallel()

		query := &Query{
			RemoteAddr: addr,
			Message: &Message{
				Questions: []Question{
					{Name: "foo.mx.", Type: TypeMX},
				},
			},
		}

		msg, err := client.Do(context.Background(), query)
		if err != nil {
			t.Fatal(err)
		}
		if want, got := len(mailZone.RRs["foo"]), len(msg.Answers); want != got {
			t.Fatalf("want %d answers, got %d", want, got)
		}
		for i, rec := range mailZone.RRs["foo"] {
			if want, got := rec, msg.Answers[i].Record; !reflect.DeepEqual(want, got) {
				t.Errorf("want MX record %#v, got %#v", want, got)
			}
		}
	})

	t.Run("two MX question query", func(t *testing.T) {
		t.Parallel()

		query := &Query{
			RemoteAddr: addr,
			Message: &Message{
				Questions: []Question{
					{Name: "foo.mx.", Type: TypeMX},
					{Name: "bar.mx.", Type: TypeMX},
				},
			},
		}

		msg, err := client.Do(context.Background(), query)
		if err != nil {
			t.Fatal(err)
		}
		if want, got := len(mailZone.RRs["foo"])+len(mailZone.RRs["bar"]), len(msg.Answers); want != got {
			t.Fatalf("want %d answers, got %d", want, got)
		}

		for i, rec := range append(mailZone.RRs["foo"], mailZone.RRs["bar"]...) {
			if want, got := rec, msg.Answers[i].Record; !reflect.DeepEqual(want, got) {
				t.Errorf("want MX record %#v, got %#v", want, got)
			}
		}
	})

	t.Run("localhost zone query", func(t *testing.T) {
		t.Parallel()

		query := &Query{
			RemoteAddr: addr,
			Message: &Message{
				Questions: []Question{
					{Name: "app.localhost.", Type: TypeA},
					{Name: "app.localhost.", Type: TypeAAAA},
				},
			},
		}

		msg, err := client.Do(context.Background(), query)
		if err != nil {
			t.Fatal(err)
		}
		if want, got := len(localhostZone.RRs["app"]), len(msg.Answers); want != got {
			t.Fatalf("want %d answers, got %d", want, got)
		}

		for i, rec := range localhostZone.RRs["app"] {
			if want, got := rec, msg.Answers[i].Record; !reflect.DeepEqual(want, got) {
				t.Errorf("want localhost record %#v, got %#v", want, got)
			}
		}
	})

	t.Run("forwarded questions query", func(t *testing.T) {
		t.Parallel()

		query := &Query{
			RemoteAddr: addr,
			Message: &Message{
				Questions: []Question{
					{Name: "test.local.", Type: TypeA},
					{Name: "test.local.", Type: TypeAAAA},
				},
			},
		}

		msg, err := client.Do(context.Background(), query)
		if err != nil {
			t.Fatal(err)
		}
		if want, got := Refused, msg.RCode; want != got {
			t.Errorf("want response RCODE %d, got %d", want, got)
		}
	})
}
