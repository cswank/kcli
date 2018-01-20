package dns

import (
	"context"
	cryptorand "crypto/rand"
	"io"
	"math/big"
	"net"
	"sync/atomic"
)

// NameServers is a slice of DNS nameserver addresses.
type NameServers []net.Addr

// Random picks a random Addr from s every time.
func (s NameServers) Random(rand io.Reader) ProxyFunc {
	max := big.NewInt(int64(len(s)))
	return func(_ context.Context, _ net.Addr) (net.Addr, error) {
		idx, err := cryptorand.Int(rand, max)
		if err != nil {
			return nil, err
		}

		return s[idx.Uint64()], nil
	}
}

// RoundRobin picks the next Addr of s by index of the last pick.
func (s NameServers) RoundRobin() ProxyFunc {
	var idx uint32
	return func(_ context.Context, _ net.Addr) (net.Addr, error) {
		return s[int(atomic.AddUint32(&idx, 1)-1)%len(s)], nil
	}
}
