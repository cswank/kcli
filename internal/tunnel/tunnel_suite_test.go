package tunnel_test

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"testing"

	"github.com/cswank/kcli/internal/tunnel"
	"github.com/gliderlabs/ssh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTunnel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tunnel Suite")
}

var _ = Describe("Tunnel", func() {

	var (
		tun      *tunnel.Tunnel
		listener *MemoryListener
		listen   func(string, string) (net.Listener, error)
		addrs    []string
	)

	BeforeEach(func() {
		ssh.Handle(func(s ssh.Session) {
			fmt.Println("handling")
			buf := []byte{}
			n, err := s.Read(buf)
			fmt.Println(n, err, string(buf))
			fmt.Println("user", s.User())
			io.WriteString(s, "Hello world\n")
		})

		go func() {
			Expect(ssh.ListenAndServe("127.0.0.1:9999", nil)).To(BeNil())
		}()

		listener = newMemoryListener()
		listen = func(protocol, host string) (net.Listener, error) {
			return listener, nil
		}

		addrs = []string{"127.0.0.1:9998", "localhost:9999"}
		tun = tunnel.New("me", 22, addrs, tunnel.Listen(listen))
	})

	Describe("Connect", func() {

		It("connects", func() {
			Expect(tun.Connect()).To(BeNil())
			conn, err := net.Dial("tcp", "127.0.0.1:9999")
			Expect(err).To(BeNil())
			fmt.Fprintf(conn, "hello server\n")
			// listen for reply
			message, err := bufio.NewReader(conn).ReadString('\n')
			Expect(err).To(BeNil())
			fmt.Println("Message from server: " + message)
		})
	})
})

//Thanks https://github.com/hydrogen18/memlistener
type MemoryListener struct {
	connections   chan net.Conn
	state         chan int
	isStateClosed uint32
}

func newMemoryListener() *MemoryListener {
	ml := &MemoryListener{}
	ml.connections = make(chan net.Conn)
	ml.state = make(chan int)
	return ml
}

func (ml *MemoryListener) Accept() (net.Conn, error) {
	select {
	case newConnection := <-ml.connections:
		return newConnection, nil
	case <-ml.state:
		return nil, errors.New("Listener closed")
	}
}

func (ml *MemoryListener) Close() error {
	if atomic.CompareAndSwapUint32(&ml.isStateClosed, 0, 1) {
		close(ml.state)
	}
	return nil
}

func (ml *MemoryListener) Dial(network, addr string) (net.Conn, error) {
	select {
	case <-ml.state:
		return nil, errors.New("Listener closed")
	default:
	}
	//Create an in memory transport
	serverSide, clientSide := net.Pipe()
	//Pass half to the server
	ml.connections <- serverSide
	//Return the other half to the client
	return clientSide, nil

}

type memoryAddr int

func (memoryAddr) Network() string {
	return "memory"
}

func (memoryAddr) String() string {
	return "local"
}
func (ml *MemoryListener) Addr() net.Addr {
	return memoryAddr(0)
}
