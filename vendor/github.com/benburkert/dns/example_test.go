package dns_test

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/benburkert/dns"
)

func ExampleClient_overrideNameServers() {
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,

		Dial: (&dns.Client{
			Transport: &dns.Transport{
				Proxy: dns.NameServers{
					&net.UDPAddr{IP: net.IPv4(8, 8, 8, 8), Port: 53},
					&net.UDPAddr{IP: net.IPv4(8, 8, 4, 4), Port: 53},
				}.RoundRobin(),
			},
		}).Dial,
	}

	addrs, err := net.LookupHost("127.0.0.1.xip.io")
	if err != nil {
		panic(err)
	}

	for _, addr := range addrs {
		fmt.Println(addr)
	}
	// Output: 127.0.0.1
}

func ExampleClient_dnsOverTLS() {
	dnsLocal := dns.OverTLSAddr{
		Addr: &net.TCPAddr{
			IP:   net.IPv4(192, 168, 8, 8),
			Port: 853,
		},
	}

	client := &dns.Client{
		Transport: &dns.Transport{
			Proxy: dns.NameServers{dnsLocal}.Random(rand.Reader),

			TLSConfig: &tls.Config{
				ServerName: "dns.local",
			},
		},
	}

	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial:     client.Dial,
	}
}

func ExampleServer_authoritative() {
	customTLD := &dns.Zone{
		Origin: "tld.",
		TTL:    time.Hour,
		SOA: &dns.SOA{
			NS:     "dns.tld.",
			MBox:   "hostmaster.tld.",
			Serial: 1234,
		},
		RRs: map[string][]dns.Record{
			"1.app": {
				&dns.A{A: net.IPv4(10, 42, 0, 1).To4()},
				&dns.AAAA{AAAA: net.ParseIP("dead:beef::1")},
			},
			"2.app": {
				&dns.A{A: net.IPv4(10, 42, 0, 2).To4()},
				&dns.AAAA{AAAA: net.ParseIP("dead:beef::2")},
			},
			"3.app": {
				&dns.A{A: net.IPv4(10, 42, 0, 3).To4()},
				&dns.AAAA{AAAA: net.ParseIP("dead:beef::3")},
			},
			"app": {
				&dns.A{A: net.IPv4(10, 42, 0, 1).To4()},
				&dns.A{A: net.IPv4(10, 42, 0, 2).To4()},
				&dns.A{A: net.IPv4(10, 42, 0, 3).To4()},
				&dns.AAAA{AAAA: net.ParseIP("dead:beef::1")},
				&dns.AAAA{AAAA: net.ParseIP("dead:beef::2")},
				&dns.AAAA{AAAA: net.ParseIP("dead:beef::3")},
			},
		},
	}

	srv := &dns.Server{
		Addr:    ":53351",
		Handler: customTLD,
	}

	go srv.ListenAndServe(context.Background())
	time.Sleep(100 * time.Millisecond) // wait for bind()

	addr, err := net.ResolveTCPAddr("tcp", srv.Addr)
	if err != nil {
		log.Fatal(err)
	}

	query := &dns.Query{
		RemoteAddr: addr,
		Message: &dns.Message{
			Questions: []dns.Question{
				{
					Name:  "app.tld.",
					Type:  dns.TypeA,
					Class: dns.ClassIN,
				},
				{
					Name:  "app.tld.",
					Type:  dns.TypeAAAA,
					Class: dns.ClassIN,
				},
			},
		},
	}

	res, err := new(dns.Client).Do(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}

	for _, answer := range res.Answers {
		switch rec := answer.Record.(type) {
		case *dns.A:
			fmt.Println(rec.A)
		case *dns.AAAA:
			fmt.Println(rec.AAAA)
		default:
			fmt.Println(rec)
		}
	}

	// Output: 10.42.0.1
	// 10.42.0.2
	// 10.42.0.3
	// dead:beef::1
	// dead:beef::2
	// dead:beef::3
}

func ExampleServer_recursive() {
	srv := &dns.Server{
		Addr:    ":53352",
		Handler: dns.HandlerFunc(dns.Recursor),
		Forwarder: &dns.Client{
			Transport: &dns.Transport{
				Proxy: dns.NameServers{
					&net.UDPAddr{IP: net.IPv4(8, 8, 8, 8), Port: 53},
					&net.UDPAddr{IP: net.IPv4(8, 8, 4, 4), Port: 53},
				}.RoundRobin(),
			},
			Resolver: new(dns.Cache),
		},
	}

	go srv.ListenAndServe(context.Background())
	time.Sleep(100 * time.Millisecond) // wait for bind()

	addr, err := net.ResolveTCPAddr("tcp", srv.Addr)
	if err != nil {
		log.Fatal(err)
	}

	query := &dns.Query{
		RemoteAddr: addr,
		Message: &dns.Message{
			RecursionDesired: true,
			Questions: []dns.Question{
				{
					Name:  "127.1.2.3.xip.io.",
					Type:  dns.TypeA,
					Class: dns.ClassIN,
				},
			},
		},
	}

	res, err := new(dns.Client).Do(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}

	for _, answer := range res.Answers {
		switch rec := answer.Record.(type) {
		case *dns.A:
			fmt.Println(rec.A)
		default:
			fmt.Println(rec)
		}
	}

	// Output: 127.1.2.3
}

func ExampleServer_recursiveWithZone() {
	customTLD := &dns.Zone{
		Origin: "tld.",
		RRs: map[string][]dns.Record{
			"foo": {
				&dns.A{A: net.IPv4(127, 0, 0, 1).To4()},
			},
		},
	}

	mux := new(dns.ResolveMux)
	mux.Handle(dns.TypeANY, "tld.", customTLD)

	srv := &dns.Server{
		Addr:    ":53353",
		Handler: mux,
		Forwarder: &dns.Client{
			Transport: &dns.Transport{
				Proxy: dns.NameServers{
					&net.UDPAddr{IP: net.IPv4(8, 8, 8, 8), Port: 53},
					&net.UDPAddr{IP: net.IPv4(8, 8, 4, 4), Port: 53},
				}.RoundRobin(),
			},
			Resolver: new(dns.Cache),
		},
	}

	go srv.ListenAndServe(context.Background())
	time.Sleep(100 * time.Millisecond) // wait for bind()

	addr, err := net.ResolveTCPAddr("tcp", srv.Addr)
	if err != nil {
		log.Fatal(err)
	}

	query := &dns.Query{
		RemoteAddr: addr,
		Message: &dns.Message{
			RecursionDesired: true,
			Questions: []dns.Question{
				{
					Name:  "127.0.0.127.xip.io.",
					Type:  dns.TypeA,
					Class: dns.ClassIN,
				},
				{
					Name:  "foo.tld.",
					Type:  dns.TypeA,
					Class: dns.ClassIN,
				},
			},
		},
	}

	res, err := new(dns.Client).Do(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}

	for _, answer := range res.Answers {
		switch rec := answer.Record.(type) {
		case *dns.A:
			fmt.Println(rec.A)
		default:
			fmt.Println(rec)
		}
	}

	// Output: 127.0.0.127
	// 127.0.0.1
}
