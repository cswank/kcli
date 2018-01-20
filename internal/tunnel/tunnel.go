package tunnel

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/benburkert/dns"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type Endpoint struct {
	Host string
	Port int
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

type connection struct {
	local  string
	server string
	remote string

	cfg *ssh.ClientConfig
}

func Connect(user string, addrs []string) error {
	out := make([]string, len(addrs))
	zone := &dns.Zone{
		TTL: 5 * time.Minute,
		RRs: map[string][]dns.Record{},
	}

	for i, addr := range addrs {
		host := strings.Split(addr, ":")[0]
		zone.RRs[host] = []dns.Record{
			&dns.A{A: net.IPv4(127, 0, 0, 1).To4()},
			&dns.AAAA{AAAA: net.ParseIP("::1")},
		}

		c := connection{
			server: fmt.Sprintf("%s:22", host),
			remote: addr,
			local:  "127.0.0.1:",
			cfg: &ssh.ClientConfig{
				User: user,
				Auth: []ssh.AuthMethod{
					sshAgent(),
				},
				HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
					return nil
				},
			},
		}
		local, err := c.start()
		if err != nil {
			return err
		}
		out[i] = local
	}

	srv := &dns.Server{
		Addr:    ":53",
		Handler: zone,
	}

	go srv.ListenAndServe(context.Background())

	mux := new(dns.ResolveMux)
	mux.Handle(dns.TypeANY, zone.Origin, zone)

	client := &dns.Client{
		Resolver: mux,
	}

	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial:     client.Dial,
	}

	return nil
}

func Close() error {
	return nil
}

func (c *connection) start() (string, error) {
	listener, err := net.Listen("tcp", c.local)
	if err != nil {
		return "", err
	}
	//defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		serverConn, err := ssh.Dial("tcp", c.server, c.cfg)
		if err != nil {
			log.Fatal(err)
		}

		remoteConn, err := serverConn.Dial("tcp", c.remote)
		log.Println("remote conn", c.remote, err)
		if err != nil {
			log.Fatal(err)
		}

		copyConn := func(writer, reader net.Conn) {
			log.Println("io copy", err)
			_, err := io.Copy(writer, reader)
			log.Println("io copy", err)
			if err != nil {
				log.Fatalf("io.Copy error: %s", err)
			}
		}

		go copyConn(conn, remoteConn)
		go copyConn(remoteConn, conn)
	}()
	return listener.Addr().String(), nil
}

func sshAgent() ssh.AuthMethod {
	if a, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(a).Signers)
	}
	return nil
}
