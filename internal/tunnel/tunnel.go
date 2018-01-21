package tunnel

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
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
	local     string
	sshServer string
	remote    string

	cfg *ssh.ClientConfig
}

//Connect creates an ssh tunnel for each address
//passed in.  It also creates a dns lookup for the local
//net.DefaultResolver so that hostname resolves to
//the localhost address of the ssh tunnel.
func Connect(user string, sshPort int, addrs []string) error {
	out := make([]string, len(addrs))
	zone := &dns.Zone{
		TTL: 5 * time.Minute,
		RRs: map[string][]dns.Record{},
	}

	for i, addr := range addrs {
		local, host, err := doConnect(addr, user, sshPort)
		if err != nil {
			return err
		}

		zone.RRs[host] = localDNS()
		out[i] = local
	}

	resolve(zone)
	return nil
}

func localDNS() []dns.Record {
	return []dns.Record{
		&dns.A{A: net.IPv4(127, 0, 0, 1).To4()},
		&dns.AAAA{AAAA: net.ParseIP("::1")},
	}
}

func resolve(zone *dns.Zone) {
	mux := new(dns.ResolveMux)
	mux.Handle(dns.TypeANY, zone.Origin, zone)

	client := &dns.Client{
		Resolver: mux,
	}

	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial:     client.Dial,
	}
}

//doConnect creates and starts a single ssh tunnel
func doConnect(addr, user string, sshPort int) (string, string, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return "", "", err
	}

	host := u.Hostname()
	c := connection{
		sshServer: fmt.Sprintf("%s:%d", host, sshPort),
		remote:    addr,
		local:     "127.0.0.1:",
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
		return "", "", err
	}

	return local, host, nil
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

	go func(listener net.Listener) {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		serverConn, err := ssh.Dial("tcp", c.sshServer, c.cfg)
		if err != nil {
			log.Fatal(err)
		}

		remoteConn, err := serverConn.Dial("tcp", c.remote)
		if err != nil {
			log.Fatal(err)
		}

		copyConn := func(writer, reader net.Conn) {
			_, err := io.Copy(writer, reader)
			if err != nil {
				log.Fatalf("io.Copy error: %s", err)
			}
		}

		go copyConn(conn, remoteConn)
		go copyConn(remoteConn, conn)
	}(listener)
	return listener.Addr().String(), nil
}

func sshAgent() ssh.AuthMethod {
	if a, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(a).Signers)
	}
	return nil
}
