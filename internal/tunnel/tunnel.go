package tunnel

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

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

func Connect(user string, addrs []string) ([]string, error) {
	out := make([]string, len(addrs))
	for i, addr := range addrs {
		c := connection{
			server: fmt.Sprintf("%s:22", strings.Split(addr, ":")[0]),
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
			return nil, err
		}
		out[i] = local
	}
	return out, nil
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
		log.Println("connect", conn, err)
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
