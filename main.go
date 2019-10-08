package main

import (
	"io/ioutil"
	"log"
	"os"
	"plugin"
	"strings"

	"github.com/cswank/kcli/internal/kafka"
	"github.com/cswank/kcli/internal/views"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	addrs     = kingpin.Flag("addresses", "comma separated list of kafka addresses").Default("localhost:9092").Short('a').Strings()
	logout    = kingpin.Flag("log", "for debugging, set the log output to a file").Short('l').String()
	topic     = kingpin.Flag("topic", "go directly to a topic").Short('t').String()
	partition = kingpin.Flag("partition", "go directly to a partition of a topic").Short('p').Default("-1").Int()
	offset    = kingpin.Flag("offset", "go directly to a message").Short('o').Default("-1").Int()
	decoder   = kingpin.Flag("decoder", "path to a plugin to decode kafka messages").Short('d').String()
	f         *os.File
)

func init() {
	kingpin.Parse()
}

func main() {
	cli := connect()
	setLogout()
	err := views.NewGui(cli, *topic, *partition, *offset)
	if f != nil {
		f.Close()
		log.SetOutput(os.Stderr)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func connect() *kafka.Client {
	var opts []kafka.Opt
	if *decoder != "" {
		dec := getDecoder(*decoder)
		opts = []kafka.Opt{kafka.WithDecoder(dec)}
	}

	cli, err := kafka.New(getAddresses(*addrs), opts...)
	if err != nil {
		log.Fatal(err)
	}

	return cli
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

func getAddresses(addrs []string) []string {
	var out []string
	for _, addr := range addrs {
		out = append(out, strings.Split(addr, ",")...)
	}

	return out
}

func getDecoder(pth string) kafka.Decoder {
	plug, err := plugin.Open(pth)
	if err != nil {
		log.Fatal(err)
	}

	s, err := plug.Lookup("Decoder")
	if err != nil {
		log.Fatal(err)
	}

	var dec kafka.Decoder
	dec, ok := s.(kafka.Decoder)
	if !ok {
		log.Fatalf("unexpected type from module symbol")
	}

	return dec
}
