package dns

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestQuestionPackUnpack(t *testing.T) {
	t.Parallel()

	tests := []struct {
		question Question
		raw      []byte
	}{
		{
			question: Question{
				Name:  ".",
				Type:  TypeA,
				Class: ClassIN,
			},

			raw: []byte{0x0, 0x0, 0x1, 0x0, 0x1},
		},
		{
			question: Question{
				Name:  "google.com.",
				Type:  TypeAAAA,
				Class: ClassIN,
			},

			raw: []byte{
				0x6, 'g', 'o', 'o', 'g', 'l', 'e',
				0x3, 'c', 'o', 'm',
				0x0,
				0x0, 0x1C, 0x0, 0x1,
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(string(test.question.Name), func(t *testing.T) {
			t.Parallel()

			raw, err := test.question.Pack(nil, nil)
			if err != nil {
				t.Fatal(err)
			}

			if want, got := test.raw, raw; !bytes.Equal(want, got) {
				t.Errorf("want raw question %x, got %x", want, got)
			}

			q := new(Question)
			buf, err := q.Unpack(raw, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(buf) > 0 {
				t.Errorf("left-over data after unpack: %x", buf)
			}

			if want, got := test.question, *q; want != got {
				t.Errorf("want question %+v, got %+v", want, got)
			}
		})
	}
}

func TestNamePackUnpack(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  []byte
		err  error
	}{
		{".", []byte{0x0}, nil},
		{"google..com", nil, errZeroSegLen},
		{"google.com.", rawGoogleCom, nil},
		{".google.com.", nil, errZeroSegLen},
		{"www..google.com.", nil, errZeroSegLen},
		{"www.google.com.", append([]byte{0x3, 'w', 'w', 'w'}, rawGoogleCom...), nil},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			raw, err := compressor{}.Pack(nil, test.name)
			if err != nil {
				if want, got := test.err, err; want != got {
					t.Errorf("want err %q, got %q", want, got)
				}
				return
			}

			if want, got := test.raw, raw; !bytes.Equal(want, got) {
				t.Fatalf("want raw name %x, got %x", want, got)
			}

			name, buf, err := decompressor(nil).Unpack(raw)
			if err != nil {
				if want, got := test.err, err; want != got {
					t.Fatalf("want err %q, got %q", want, got)
				}
				return
			}
			if len(buf) > 0 {
				t.Errorf("left-over data after unpack: %x", buf)
			}

			if want, got := test.name, name; want != got {
				t.Errorf("want unpacked name %q, got %q", want, got)
			}
		})
	}
}

func TestMessagePackUnpack(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string

		msg      Message
		compress bool
		buf      []byte

		raw []byte
	}{
		{
			name: ".	IN	AAAA",

			msg: Message{
				ID:               0x1001,
				RecursionDesired: true,
				Questions: []Question{
					{
						Name:  ".",
						Type:  TypeAAAA,
						Class: ClassIN,
					},
				},
			},

			raw: []byte{
				0x10, 0x01, // ID=0x1001
				0x01, 0x00, // RD=1
				0x00, 0x01, // QDCOUNT=1
				0x00, 0x00, // ANCOUNT=0
				0x00, 0x00, // NSCOUNT=0
				0x00, 0x00, // ARCOUNT=0

				0x00, 0x00, 0x1C, 0x00, 0x01, // .	IN	AAAA
			},
		},
		{
			name: "txt.example.com.	IN	TXT",

			msg: Message{
				ID:       0x01,
				Response: true,
				Questions: []Question{
					{
						Name:  "txt.example.com.",
						Type:  TypeTXT,
						Class: ClassIN,
					},
				},
				Answers: []Resource{
					{
						Name:   "txt.example.com.",
						Class:  ClassIN,
						TTL:    60 * time.Second,
						Record: &TXT{[]string{"multi", "segment txt", "record"}},
					},
				},
			},

			raw: []byte{
				0x00, 0x01, // ID=0x0001
				0x80, 0x00, // QR=1
				0x00, 0x01, // QDCOUNT=1
				0x00, 0x01, // ANCOUNT=1
				0x00, 0x00, // NSCOUNT=0
				0x00, 0x00, // ARCOUNT=0

				// txt.example.com.	IN	TXT
				0x03, 't', 'x', 't',
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0x03, 'c', 'o', 'm',
				0x00,
				0x00, 0x10, 0x00, 0x01, // TYPE=TXT,CLASS=IN

				// txt.example.com.	60	IN	TXT	"abcd"
				0x03, 't', 'x', 't',
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0x03, 'c', 'o', 'm',
				0x00,
				0x00, 0x10, 0x00, 0x01, // TYPE=TXT,CLASS=IN
				0x00, 0x00, 0x00, 0x3C, // TTL=60
				0x00, 0x19, // RDLENGTH=25

				0x05, 'm', 'u', 'l', 't', 'i',
				0x0B, 's', 'e', 'g', 'm', 'e', 'n', 't', ' ', 't', 'x', 't',
				0x06, 'r', 'e', 'c', 'o', 'r', 'd',
			},
		},
		{
			name: ".	60	IN	A",

			msg: Message{
				ID:       0x100,
				Response: true,
				Questions: []Question{
					{
						Name:  ".",
						Type:  TypeA,
						Class: ClassIN,
					},
				},
				Answers: []Resource{
					{
						Name:  ".",
						Class: ClassIN,
						TTL:   60 * time.Second,
						Record: &A{
							A: net.IPv4(127, 0, 0, 1).To4(),
						},
					},
				},
			},

			raw: []byte{
				0x01, 0x00, // ID=0x0100
				0x80, 0x00, // RD=1
				0x00, 0x01, // QDCOUNT=1
				0x00, 0x01, // ANCOUNT=1
				0x00, 0x00, // NSCOUNT=0
				0x00, 0x00, // ARCOUNT=0

				0x00, 0x00, 0x01, 0x00, 0x01, // .	IN	A

				// .	60	IN	127.0.0.1
				0x00,
				0x00, 0x01, 0x00, 0x01, // TYPE=A,CLASS=IN
				0x00, 0x00, 0x00, 0x3C, // TTL=60
				0x00, 0x04,

				0x7F, 0x00, 0x00, 0x01, // 127.0.0.1
			},
		},
		{
			name: ".	60	IN	AAAA",

			msg: Message{
				ID:       0x101,
				Response: true,
				Questions: []Question{
					{
						Name:  ".",
						Type:  TypeAAAA,
						Class: ClassIN,
					},
				},
				Answers: []Resource{
					{
						Name:  ".",
						Class: ClassIN,
						TTL:   60 * time.Second,
						Record: &AAAA{
							AAAA: net.ParseIP("::1"),
						},
					},
				},
			},

			raw: []byte{
				0x01, 0x01, // ID=0x0101
				0x80, 0x00, // RD=1
				0x00, 0x01, // QDCOUNT=1
				0x00, 0x01, // ANCOUNT=1
				0x00, 0x00, // NSCOUNT=0
				0x00, 0x00, // ARCOUNT=0

				0x00, 0x00, 0x1C, 0x00, 0x01, // .	IN	AAAA

				// .	60	IN	::1
				0x00,
				0x00, 0x1C, 0x00, 0x01, // TYPE=AAAA,CLASS=IN
				0x00, 0x00, 0x00, 0x3C, // TTL=60
				0x00, 0x10,

				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, // ::1
			},
		},
		{
			name: ".	60	IN	CNAME",

			msg: Message{
				ID:       0x102,
				Response: true,
				Questions: []Question{
					{
						Name:  ".",
						Type:  TypeCNAME,
						Class: ClassIN,
					},
				},
				Answers: []Resource{
					{
						Name:  ".",
						Class: ClassIN,
						TTL:   60 * time.Second,
						Record: &CNAME{
							CNAME: "tld.",
						},
					},
				},
			},

			raw: []byte{
				0x01, 0x02, // ID=0x0102
				0x80, 0x00, // RD=1
				0x00, 0x01, // QDCOUNT=1
				0x00, 0x01, // ANCOUNT=1
				0x00, 0x00, // NSCOUNT=0
				0x00, 0x00, // ARCOUNT=0

				0x00, 0x00, 0x05, 0x00, 0x01, // .	IN	CNAME

				// .	60	IN	"tld."
				0x00,
				0x00, 0x05, 0x00, 0x01, // TYPE=CNAME,CLASS=IN
				0x00, 0x00, 0x00, 0x3C, // TTL=60
				0x00, 0x05,

				0x03, 't', 'l', 'd',
				0x00,
			},
		},
		{
			name: ".	60	IN	SOA",

			msg: Message{
				ID:       0x103,
				Response: true,
				Questions: []Question{
					{
						Name:  ".",
						Type:  TypeA,
						Class: ClassIN,
					},
				},
				Authorities: []Resource{
					{
						Name:  ".",
						Class: ClassIN,
						TTL:   60 * time.Second,
						Record: &SOA{
							NS:      "ns.",
							MBox:    "mx.",
							Serial:  1234,
							Refresh: 24 * time.Hour,
							Retry:   2 * time.Hour,
							Expire:  3 * time.Minute,
							MinTTL:  4 * time.Second,
						},
					},
				},
			},

			raw: []byte{
				0x01, 0x03, // ID=0x0103
				0x80, 0x00, // RD=1
				0x00, 0x01, // QDCOUNT=1
				0x00, 0x00, // ANCOUNT=0
				0x00, 0x01, // NSCOUNT=1
				0x00, 0x00, // ARCOUNT=0

				0x00, 0x00, 0x01, 0x00, 0x01, // .	IN	A

				// .	60	IN	ns.	mx.	1234	86400	7200	180	4
				0x00,
				0x00, 0x06, 0x00, 0x01, // TYPE=SOA,CLASS=IN
				0x00, 0x00, 0x00, 0x3C, // TTL=60
				0x00, 0x1C,

				0x02, 'n', 's',
				0x00,
				0x02, 'm', 'x',
				0x00,
				0x00, 0x00, 0x04, 0xD2, // SERIAL=1234
				0x00, 0x01, 0x51, 0x80, // REFRESH=86400
				0x00, 0x00, 0x1C, 0x20, // RETRY=720
				0x00, 0x00, 0x00, 0xB4, // EXPIRE=180
				0x00, 0x00, 0x00, 0x04, // MINIMUM=4
			},
		},
		{
			name: "1.0.0.127.in-addr.arpa.	60	IN	PTR",

			msg: Message{
				ID:       0x104,
				Response: true,
				Questions: []Question{
					{
						Name:  "1.0.0.127.in-addr.arpa.",
						Type:  TypePTR,
						Class: ClassIN,
					},
				},
				Answers: []Resource{
					{
						Name:  "1.0.0.127.in-addr.arpa.",
						Class: ClassIN,
						TTL:   60 * time.Second,
						Record: &PTR{
							PTR: "localhost.",
						},
					},
				},
			},

			raw: []byte{
				0x01, 0x04, // ID=0x0104
				0x80, 0x00, // RD=1
				0x00, 0x01, // QDCOUNT=1
				0x00, 0x01, // ANCOUNT=1
				0x00, 0x00, // NSCOUNT=0
				0x00, 0x00, // ARCOUNT=0

				0x01, '1',
				0x01, '0',
				0x01, '0',
				0x03, '1', '2', '7',
				0x07, 'i', 'n', '-', 'a', 'd', 'd', 'r',
				0x04, 'a', 'r', 'p', 'a',
				0x00,

				0x00, 0x0C, 0x00, 0x01, // TYPE=PTR,CLASS=IN

				// 1.0.0.127.in-addr.arpa.	60	IN	localhost.

				0x01, '1',
				0x01, '0',
				0x01, '0',
				0x03, '1', '2', '7',
				0x07, 'i', 'n', '-', 'a', 'd', 'd', 'r',
				0x04, 'a', 'r', 'p', 'a',
				0x00,
				0x00, 0x0C, 0x00, 0x01, // TYPE=PTR,CLASS=IN
				0x00, 0x00, 0x00, 0x3C, // TTL=60
				0x00, 0x0B,

				0x09, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't',
				0x00,
			},
		},
		{
			name: ".	60	IN	MX",

			msg: Message{
				ID:       0x105,
				Response: true,
				Questions: []Question{
					{
						Name:  ".",
						Type:  TypeMX,
						Class: ClassIN,
					},
				},
				Answers: []Resource{
					{
						Name:  ".",
						Class: ClassIN,
						TTL:   60 * time.Second,
						Record: &MX{
							Pref: 101,
							MX:   "mx.",
						},
					},
				},
			},

			raw: []byte{
				0x01, 0x05, // ID=0x0105
				0x80, 0x00, // RD=1
				0x00, 0x01, // QDCOUNT=1
				0x00, 0x01, // ANCOUNT=1
				0x00, 0x00, // NSCOUNT=0
				0x00, 0x00, // ARCOUNT=0

				0x00, 0x00, 0x0F, 0x00, 0x01, // .	IN	MX

				0x00, 0x00, 0x0F, 0x00, 0x01, // TYPE=MX,CLASS=IN
				0x00, 0x00, 0x00, 0x3C, // TTL=60
				0x00, 0x06,

				0x00, 0x65,
				0x02, 'm', 'x',
				0x00,
			},
		},
		{
			name: ".	60	IN	NS",

			msg: Message{
				ID:       0x106,
				Response: true,
				Questions: []Question{
					{
						Name:  ".",
						Type:  TypeNS,
						Class: ClassIN,
					},
				},
				Answers: []Resource{
					{
						Name:  ".",
						Class: ClassIN,
						TTL:   60 * time.Second,
						Record: &NS{
							NS: "ns.",
						},
					},
				},
			},

			raw: []byte{
				0x01, 0x06, // ID=0x0106
				0x80, 0x00, // RD=1
				0x00, 0x01, // QDCOUNT=1
				0x00, 0x01, // ANCOUNT=1
				0x00, 0x00, // NSCOUNT=0
				0x00, 0x00, // ARCOUNT=0

				0x00, 0x00, 0x02, 0x00, 0x01, // .	IN	NS

				0x00, 0x00, 0x02, 0x00, 0x01, // TYPE=NS,CLASS=IN
				0x00, 0x00, 0x00, 0x3C, // TTL=60
				0x00, 0x04,

				0x02, 'n', 's',
				0x00,
			},
		},
		{
			name: ".	60	IN	SRV",

			msg: Message{
				ID:       0x107,
				Response: true,
				Questions: []Question{
					{
						Name:  ".",
						Type:  TypeSRV,
						Class: ClassIN,
					},
				},
				Answers: []Resource{
					{
						Name:  ".",
						Class: ClassIN,
						TTL:   60 * time.Second,
						Record: &SRV{
							Priority: 0x01,
							Weight:   0x10,
							Port:     0x11,
							Target:   "srv.",
						},
					},
				},
			},

			raw: []byte{
				0x01, 0x07, // ID=0x0107
				0x80, 0x00, // RD=1
				0x00, 0x01, // QDCOUNT=1
				0x00, 0x01, // ANCOUNT=1
				0x00, 0x00, // NSCOUNT=0
				0x00, 0x00, // ARCOUNT=0

				0x00, 0x00, 0x21, 0x00, 0x01, // .	IN	SRV

				0x00, 0x00, 0x21, 0x00, 0x01, // TYPE=SRV,CLASS=IN
				0x00, 0x00, 0x00, 0x3C, // TTL=60
				0x00, 0x0B,

				0x00, 0x01, 0x00, 0x10, 0x00, 0x11,
				0x03, 's', 'r', 'v',
				0x00,
			},
		},
		{
			name: "compressed response",

			msg: Message{
				Response: true,
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
			compress: true,
			buf:      make([]byte, 2),

			raw: []byte{
				0x00, 0x00, // ID=0x0001
				0x80, 0x00, // QR=1
				0x00, 0x01, // QDCOUNT=1
				0x00, 0x01, // ANCOUNT=1
				0x00, 0x00, // NSCOUNT=0
				0x00, 0x00, // ARCOUNT=0

				// example.com.	IN	A
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0x03, 'c', 'o', 'm',
				0x00,
				0x00, 0x01, 0x00, 0x01, // TYPE=A,CLASS=IN

				// example.com	60	IN A 127.0.0.1
				0xC0, 0x0C,
				0x00, 0x01, 0x00, 0x01, // TYPE=TXT,CLASS=IN
				0x00, 0x00, 0x00, 0x3C, // TTL=60
				0x00, 0x04, // RDLENGTH=5

				0x7F, 0x00, 0x00, 0x01, // 127.0.0.1
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(string(test.name), func(t *testing.T) {
			t.Parallel()

			raw, err := test.msg.Pack(test.buf, test.compress)
			if err != nil {
				t.Fatal(err)
			}

			if want, got := append(test.buf, test.raw...), raw; !bytes.Equal(want, got) {
				t.Errorf("want raw message %+v, got %+v", want, got)
			}

			msg := new(Message)
			buf, err := msg.Unpack(raw[len(test.buf):])
			if err != nil {
				t.Fatal(err)
			}
			if len(buf) > 0 {
				t.Errorf("left-over data after unpack: %x", buf)
			}

			if want, got := test.msg, *msg; !reflect.DeepEqual(want, got) {
				t.Errorf("want message %+v, got %+v", want, got)
			}
		})
	}
}

func TestMessageCompress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  Message
		raw  []byte
	}{
		{
			name: "multi-A-question",

			msg: Message{
				Questions: []Question{
					{
						Name:  "aaa.",
						Type:  TypeA,
						Class: ClassIN,
					},
					{
						Name:  "bbb.aaa.",
						Type:  TypeA,
						Class: ClassIN,
					},
					{
						Name:  "ccc.bbb.aaa.",
						Type:  TypeA,
						Class: ClassIN,
					},
				},
			},

			raw: []byte{
				0x00, 0x00, // ID=0x0000
				0x00, 0x00, // QR=0
				0x00, 0x03, // QDCOUNT=0
				0x00, 0x00, // ANCOUNT=0
				0x00, 0x00, // NSCOUNT=0
				0x00, 0x00, // ARCOUNT=0

				// aaa.	IN	A
				0x03, 'a', 'a', 'a',
				0x00,
				0x00, 0x01, 0x00, 0x01,

				// bbb.aaa.	IN	A
				0x03, 'b', 'b', 'b',
				0xC0, 0x0C,
				0x00, 0x01, 0x00, 0x01,

				// ccc.bbb.aaa.	IN	A
				0x03, 'c', 'c', 'c',
				0xC0, 0x15,
				0x00, 0x01, 0x00, 0x01,
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(string(test.name), func(t *testing.T) {
			t.Parallel()

			raw, err := test.msg.Pack(nil, true)
			if err != nil {
				t.Fatal(err)
			}

			if want, got := test.raw, raw; !bytes.Equal(want, got) {
				t.Errorf("want raw message %+v, got %+v", want, got)
			}

			msg := new(Message)
			buf, err := msg.Unpack(raw)
			if err != nil {
				t.Fatal(err)
			}
			if len(buf) > 0 {
				t.Errorf("left-over data after unpack: %x", buf)
			}

			if want, got := test.msg, *msg; !reflect.DeepEqual(want, got) {
				t.Errorf("want message %+v, got %+v", want, got)
			}
		})
	}
}

var (
	rawGoogleCom = []byte{
		0x6, 'g', 'o', 'o', 'g', 'l', 'e',
		0x3, 'c', 'o', 'm',
		0x0,
	}
)

func BenchmarkMessagePack(b *testing.B) {
	b.Run("small-message", func(b *testing.B) {
		msg := smallTestMsg()

		for _, bufsize := range []int{0, 512} {
			bufsize := bufsize

			b.Run(fmt.Sprintf("buf=%d", bufsize), func(b *testing.B) {
				benchamarkMessagePack(b, msg, make([]byte, bufsize))
			})
		}
	})

	b.Run("large-message", func(b *testing.B) {
		msg := largeTestMsg()

		for _, bufsize := range []int{0, 512, 4096} {
			bufsize := bufsize

			b.Run(fmt.Sprintf("buf=%d", bufsize), func(b *testing.B) {
				benchamarkMessagePack(b, msg, make([]byte, bufsize))
			})
		}
	})
}

func benchamarkMessagePack(b *testing.B, msg Message, buf []byte) {
	tmp, err := msg.Pack(nil, false)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(tmp)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := msg.Pack(buf[:0], false); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageCompress(b *testing.B) {
	b.Run("small-message", func(b *testing.B) {
		msg := smallTestMsg()

		for _, bufsize := range []int{0, 512} {
			bufsize := bufsize

			b.Run(fmt.Sprintf("buf=%d", bufsize), func(b *testing.B) {
				benchamarkMessageCompress(b, msg, make([]byte, bufsize))
			})
		}
	})

	b.Run("large-message", func(b *testing.B) {
		msg := largeTestMsg()

		for _, bufsize := range []int{0, 512, 4096} {
			bufsize := bufsize

			b.Run(fmt.Sprintf("buf=%d", bufsize), func(b *testing.B) {
				benchamarkMessageCompress(b, msg, make([]byte, bufsize))
			})
		}
	})
}

func benchamarkMessageCompress(b *testing.B, msg Message, buf []byte) {
	tmp, err := msg.Pack(nil, false)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(tmp)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := msg.Pack(buf[:0], true); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageUnpack(b *testing.B) {
	b.Run("small-message", func(b *testing.B) {
		benchamarkMessageUnpack(b, smallTestMsg(), false)
	})

	b.Run("large-message", func(b *testing.B) {
		benchamarkMessageUnpack(b, largeTestMsg(), false)
	})
}

func BenchmarkMessageDecompress(b *testing.B) {
	b.Run("small-message", func(b *testing.B) {
		benchamarkMessageUnpack(b, smallTestMsg(), true)
	})

	b.Run("large-message", func(b *testing.B) {
		benchamarkMessageUnpack(b, largeTestMsg(), true)
	})
}

func benchamarkMessageUnpack(b *testing.B, msg Message, compress bool) {
	buf, err := msg.Pack(nil, compress)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(buf)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg Message
		if _, err := msg.Unpack(buf); err != nil {
			b.Fatal(err)
		}
	}
}

func smallTestMsg() Message {
	name := "example.com."

	return Message{
		Response:      true,
		Authoritative: true,

		Questions: []Question{
			{
				Name:  name,
				Type:  TypeA,
				Class: ClassIN,
			},
		},
		Answers: []Resource{
			{
				Name:  name,
				Class: ClassIN,
				Record: &A{
					A: net.IPv4(127, 0, 0, 1).To4(),
				},
			},
		},
		Authorities: []Resource{
			{
				Name:  name,
				Class: ClassIN,
				Record: &A{
					A: net.IPv4(127, 0, 0, 1).To4(),
				},
			},
		},
		Additionals: []Resource{
			{
				Name:  name,
				Class: ClassIN,
				Record: &A{
					A: net.IPv4(127, 0, 0, 1).To4(),
				},
			},
		},
	}
}

func largeTestMsg() Message {
	name := "foo.bar.example.com."

	return Message{
		Response:      true,
		Authoritative: true,
		Questions: []Question{
			{
				Name:  name,
				Type:  TypeA,
				Class: ClassIN,
			},
		},
		Answers: []Resource{
			{
				Name:  name,
				Class: ClassIN,
				Record: &A{
					A: net.IPv4(127, 0, 0, 1),
				},
			},
			{
				Name:  name,
				Class: ClassIN,
				Record: &A{
					A: net.IPv4(127, 0, 0, 2),
				},
			},
			{
				Name:  name,
				Class: ClassIN,
				Record: &AAAA{
					AAAA: net.ParseIP("::1"),
				},
			},
			{
				Name:  name,
				Class: ClassIN,
				Record: &CNAME{
					CNAME: "alias.example.com.",
				},
			},
			{
				Name:  name,
				Class: ClassIN,
				Record: &SOA{
					NS:      "ns1.example.com.",
					MBox:    "mb.example.com.",
					Serial:  1,
					Refresh: 2 * time.Second,
					Retry:   3 * time.Second,
					Expire:  4 * time.Second,
					MinTTL:  5 * time.Second,
				},
			},
			{
				Name:  name,
				Class: ClassIN,
				Record: &PTR{
					PTR: "ptr.example.com.",
				},
			},
			{
				Name:  name,
				Class: ClassIN,
				Record: &MX{
					Pref: 7,
					MX:   "mx.example.com.",
				},
			},
			{
				Name:  name,
				Class: ClassIN,
				Record: &SRV{
					Priority: 8,
					Weight:   9,
					Port:     11,
					Target:   "srv.example.com.",
				},
			},
		},
		Authorities: []Resource{
			{
				Name:  name,
				Class: ClassIN,
				Record: &NS{
					NS: "ns1.example.com.",
				},
			},
			{
				Name:  name,
				Class: ClassIN,
				Record: &NS{
					NS: "ns2.example.com.",
				},
			},
		},
		Additionals: []Resource{
			{
				Name:  name,
				Class: ClassIN,
				Record: &TXT{
					TXT: []string{"So Long, and Thanks for All the Fish"},
				},
			},
			{
				Name:  name,
				Class: ClassIN,
				Record: &TXT{
					TXT: []string{"Hamster Huey and the Gooey Kablooie"},
				},
			},
		},
	}
}
