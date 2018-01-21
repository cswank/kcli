package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/cswank/kcli/internal/kafka"
	"github.com/cswank/kcli/internal/tunnel"
	"github.com/cswank/kcli/internal/views"

	ui "github.com/jroimartin/gocui"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	g         *ui.Gui
	addrs     = kingpin.Flag("addresses", "comma seperated list of kafka addresses").Default("localhost:9092").Short('a').Strings()
	logout    = kingpin.Flag("log", "for debugging, set the log output to a file").Short('l').String()
	topic     = kingpin.Flag("topic", "go directly to a topic").Short('t').String()
	partition = kingpin.Flag("partition", "go directly to a partition of a topic").Short('p').Default("-1").Int()
	offset    = kingpin.Flag("offset", "go directly to a message").Short('o').Default("-1").Int()
	ssh       = kingpin.Flag("ssh", "ssh username for tunneling to kafka hosts").String()
	sshPort   = kingpin.Flag("port", "ssh port for tunneling to kafka hosts").Default("22").Int()

	f *os.File
)

func init() {
	kingpin.Parse()
}

func connect() {
	a, err := getAddresses(*addrs)
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	if err := kafka.Connect(a); err != nil {
		log.Fatal(err)
	}
}

func main() {
	setLogout()
	connect()
	err := views.NewGui(*topic, *partition, *offset)
	if f != nil {
		f.Close()
		log.SetOutput(os.Stderr)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func setLogout() {
	if *logout != "" {
		var err error
		f, err = os.Create(*logout)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(f)
	} else {
		log.SetOutput(ioutil.Discard)
	}
}

func getAddresses(addrs []string) ([]string, error) {
	var out []string
	for _, addr := range addrs {
		out = append(out, strings.Split(addr, ",")...)
	}

	if *ssh != "" {
		if err := tunnel.Connect(*ssh, *sshPort, out); err != nil {
			return nil, err
		}
	}

	return out, nil
}
