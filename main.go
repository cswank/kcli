package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/cswank/kcli/internal/streams"
	"github.com/cswank/kcli/internal/views"

	ui "github.com/jroimartin/gocui"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	g         *ui.Gui
	addrs     = kingpin.Flag("addresses", "comma separated list of kafka addresses").Default("localhost:9092").Short('a').Strings()
	logout    = kingpin.Flag("log", "for debugging, set the log output to a file").Short('l').String()
	topic     = kingpin.Flag("topic", "go directly to a topic").Short('t').String()
	partition = kingpin.Flag("partition", "go directly to a partition of a topic").Short('p').Default("-1").Int()
	offset    = kingpin.Flag("offset", "go directly to a message").Short('o').Default("-1").Int()
	ssh       = kingpin.Flag("ssh", "ssh username for tunneling to a kafka host.  This currently only works for a single kafka node, not a cluster").String()
	sshPort   = kingpin.Flag("port", "ssh port for tunneling to kafka hosts").Default("22").Int()

	region = kingpin.Flag("region", "aws region (if set then connect to kinesis").Short('r').String()

	f *os.File
)

func init() {
	kingpin.Parse()
}

func connect() (streams.Streamer, error) {
	var cli streams.Streamer
	var err error

	if *region == "" {
		cli, err = streams.NewKafka(getAddresses(*addrs), *ssh, *sshPort)
	} else {
		cli, err = streams.NewSQS(*region)
		//cli, err = streams.NewKinesis(*region)
	}
	return cli, err
}

func main() {
	setLogout()
	cli, err := connect()
	if err != nil {
		log.Fatal(err)
	}

	err = views.NewGui(cli, *topic, *partition, *offset)
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

func getAddresses(addrs []string) []string {
	var out []string
	for _, addr := range addrs {
		out = append(out, strings.Split(addr, ",")...)
	}

	return out
}
