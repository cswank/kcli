package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/cswank/kcli/internal/kafka"
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

	f *os.File
)

func init() {
	kingpin.Parse()

	if err := kafka.Connect(getAddresses(*addrs)); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var err error
	g, err = ui.NewGui(ui.Output256)
	if err != nil {
		g.Close()
		kafka.Close()
		log.Fatal("could not create gui", err)
	}

	w, h := g.Size()

	opts := getOpts()
	s, err := views.NewScreen(g, w, h, opts...)
	if err != nil {
		g.Close()
		kafka.Close()
		log.Fatalf("error: %s", err)
	}

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

	g.SetManagerFunc(s.GetLayout(g, w, h))
	g.Cursor = true
	g.InputEsc = true

	var closed bool
	defer func() {
		if !closed {
			g.Close()
		}
		if f != nil {
			f.Close()
		}
		kafka.Close()
	}()

	if err := s.Keybindings(g); err != nil {
		log.Println(err)
		return
	}

	if err := g.MainLoop(); err != nil {
		if err != ui.ErrQuit {
			log.SetOutput(os.Stderr)
			log.Println(err)
			return
		}
	}

	closed = true
	g.Close()
	if s.After != nil {
		s.After()
	}
}

func getOpts() []views.Init {
	var out []views.Init
	if *topic != "" {
		out = append(out, views.EnterTopic(*topic))
		if *partition != -1 {
			out = append(out, views.EnterPartition(*partition))
			if *offset != -1 {
				out = append(out, views.EnterOffset(*offset))
			}
		}
	}
	return out
}

func getAddresses(addrs []string) []string {
	var out []string
	for _, addr := range addrs {
		out = append(out, strings.Split(addr, ",")...)
	}
	return out
}
