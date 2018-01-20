package dns

import (
	"context"
	"strings"
	"time"
)

// Zone is a contiguous set DNS records under an origin domain name.
type Zone struct {
	Origin string
	TTL    time.Duration

	SOA *SOA

	RRs map[string][]Record
}

// ServeDNS answers DNS queries in zone z.
func (z *Zone) ServeDNS(ctx context.Context, w MessageWriter, r *Query) {
	w.Authoritative(true)

	var found bool
	for _, q := range r.Questions {
		if !strings.HasSuffix(q.Name, z.Origin) {
			continue
		}

		dn := q.Name[:len(q.Name)-len(z.Origin)-1]

		for _, rr := range z.RRs[dn] {
			if q.Type != rr.Type() {
				continue
			}

			w.Answer(q.Name, z.TTL, rr)

			found = true
		}
	}

	if !found {
		w.Status(NXDomain)

		if z.SOA != nil {
			w.Authority(z.Origin, z.TTL, z.SOA)
		}
	}
}
