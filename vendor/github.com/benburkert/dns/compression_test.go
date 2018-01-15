package dns

import (
	"bytes"
	"testing"
)

func TestCompressor(t *testing.T) {
	tests := []struct {
		name string

		fqdn  string
		state map[string]int
		buf   []byte

		raw []byte
		err error
	}{
		{
			name: ".",

			fqdn: ".",

			raw: []byte{0x00},
		},
		{
			name: "example.com",

			fqdn: "example.com.",

			raw: []byte{
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0x03, 'c', 'o', 'm',
				0x00,
			},
		},
		{
			name: "compressed-example.com",

			fqdn:  "example.com.",
			state: map[string]int{"com.": 5},
			buf:   make([]byte, 2),

			raw: []byte{
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0xC0, 0x05,
			},
		},
	}

	t.Parallel()

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			com := compressor{test.state, 0}

			if want, got := len(test.raw), com.Length(test.fqdn); want != got {
				t.Fatalf("want compressed length %d, got %d", want, got)
			}

			raw, err := com.Pack(test.buf, test.fqdn)
			if err != nil {
				if want, got := test.err, err; want != got {
					t.Errorf("want err %q, got %q", want, got)
				}
				return
			}

			if want, got := append(test.buf, test.raw...), raw; !bytes.Equal(want, got) {
				t.Errorf("want compressed name %x, got %x", want, got)
			}
		})
	}
}

func TestDecompressor(t *testing.T) {
	tests := []struct {
		name string

		raw   []byte
		state []byte

		fqdn string
		err  error
	}{
		{
			name: ".",

			raw: []byte{0x00},

			fqdn: ".",
		},
		{
			name: "example.com",

			raw: []byte{
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0x03, 'c', 'o', 'm',
				0x00,
			},

			fqdn: "example.com.",
		},
		{
			name: "compressed-example.com",

			raw: []byte{
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0xC0, 0x05,
			},
			state: []byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				0x03, 'c', 'o', 'm',
				0x00,
			},

			fqdn: "example.com.",
		},
	}

	t.Parallel()

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			dec := decompressor(test.state)

			fqdn, _, err := dec.Unpack(test.raw)
			if err != nil {
				if want, got := test.err, err; want != got {
					t.Errorf("want err %q, got %q", want, got)
				}
				return
			}

			if want, got := test.fqdn, fqdn; want != got {
				t.Errorf("want decompressed name %q, got %q", want, got)
			}
		})
	}
}
