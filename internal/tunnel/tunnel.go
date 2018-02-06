package tunnel

import (
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

type endpoint struct {
	Host string
	Port int
}

func (e *endpoint) String() string {
	return fmt.Sprintf("%s:%d", e.Host, e.Port)
}

type connection struct {
	local     string
	sshServer string
	remote    string
	cfg       *ssh.ClientConfig
}

//Tunnel holds data needed to create one
//ssh tunnel per host for port forwarding.
type Tunnel struct {
	user    string
	sshPort int
	addrs   []string
}

//New creates a new Tunnel
func New(user string, port int, addrs []string) *Tunnel {
	return &Tunnel{
		user:    user,
		sshPort: port,
		addrs:   addrs,
	}
}

//Connect creates an ssh tunnel for each address
//passed in.  It also creates a dns lookup for the local
//net.DefaultResolver so that hostname resolves to
//the localhost address of the ssh tunnel.
func (t *Tunnel) Connect() ([]string, error) {
	out := make([]string, len(t.addrs))
	zone := &dns.Zone{
		TTL: 5 * time.Minute,
		RRs: map[string][]dns.Record{},
	}

	for i, addr := range t.addrs {
		local, host, err := t.doConnect(addr)
		if err != nil {
			return nil, err
		}

		zone.RRs[host] = t.localDNS()
		out[i] = fmt.Sprintf("%s:%s", host, strings.Split(local, ":")[1])
	}

	t.resolve(zone)
	return out, nil
}

func (t *Tunnel) localDNS() []dns.Record {
	return []dns.Record{
		&dns.A{A: net.IPv4(127, 0, 0, 1).To4()},
		&dns.AAAA{AAAA: net.ParseIP("::1")},
	}
}

func (t *Tunnel) resolve(zone *dns.Zone) {
	mux := new(dns.ResolveMux)
	mux.Handle(dns.TypeANY, zone.Origin, zone)

	client := &dns.Client{
		Resolver: mux,
	}

	r := &net.Resolver{
		PreferGo: true,
		Dial:     client.Dial,
	}

	net.DefaultResolver = r
}

//doConnect creates and starts a single ssh tunnel
func (t *Tunnel) doConnect(addr string) (string, string, error) {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid remote address (no port): %s", addr)
	}

	host := parts[0]
	addrs, err := net.LookupHost(host)
	fmt.Println("looked up", addrs, err)
	if err != nil {
		return "", "", err
	}

	if len(addrs) == 0 {
		return "", "", fmt.Errorf("couldn't look up host %s", host)
	}

	c := connection{
		sshServer: fmt.Sprintf("%s:%d", addrs[0], t.sshPort),
		remote:    fmt.Sprintf("%s:%s", addrs[0], parts[1]),
		local:     "127.0.0.1:",
		cfg: &ssh.ClientConfig{
			User: t.user,
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
