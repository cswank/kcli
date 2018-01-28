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

type Endpoint struct {
	Host string
	Port int
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

type getListener func(protocol, host string) (net.Listener, error)

type connection struct {
	local     string
	sshServer string
	remote    string
	listen    getListener
	cfg       *ssh.ClientConfig
}

type Tunnel struct {
	user    string
	sshPort int
	addrs   []string
	listen  getListener
}

func New(user string, sshPort int, addrs []string, opts ...func(t *Tunnel)) *Tunnel {
	t := &Tunnel{
		user:    user,
		sshPort: sshPort,
		addrs:   addrs,
	}

	for _, o := range opts {
		o(t)
	}

	if t.listen == nil {
		t.listen = net.Listen
	}

	return t
}

func Listen(f getListener) func(*Tunnel) {
	return func(t *Tunnel) {
		t.listen = f
	}
}

//Connect creates an ssh tunnel for each address
//passed in.  It also creates a dns lookup for the local
//net.DefaultResolver so that hostname resolves to
//the localhost address of the ssh tunnel.
func (t *Tunnel) Connect() error {
	out := make([]string, len(t.addrs))
	zone := &dns.Zone{
		TTL: 5 * time.Minute,
		RRs: map[string][]dns.Record{},
	}

	for i, addr := range t.addrs {
		local, host, err := t.doConnect(addr)
		if err != nil {
			return err
		}

		fmt.Println("host", host)
		zone.RRs[host] = t.localDNS()
		out[i] = local
	}

	t.resolve(zone)
	return nil
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

	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial:     client.Dial,
	}
}

//doConnect creates and starts a single ssh tunnel
func (t *Tunnel) doConnect(addr string) (string, string, error) {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid remote address (no port): %s", addr)
	}

	host := parts[0]
	c := connection{
		listen:    t.listen,
		sshServer: fmt.Sprintf("%s:%d", host, t.sshPort),
		remote:    addr,
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
	fmt.Printf("starting connection %+v\n", c)
	listener, err := c.listen("tcp", c.local)
	if err != nil {
		return "", err
	}
	//defer listener.Close()

	go func(listener net.Listener) {
		conn, err := listener.Accept()
		fmt.Println("listener accepted", conn, err)
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
