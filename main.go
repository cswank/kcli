package main

import (
	"log"
	"os"

	"github.com/cswank/kcli/internal/kafka"
	"github.com/cswank/kcli/internal/views"

	ui "github.com/jroimartin/gocui"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	g      *ui.Gui
	addrs  = kingpin.Flag("addresses", "comma seperated list of kafka addresses").Default("localhost:9092").Short('a').Strings()
	logout = kingpin.Flag("logs", "for debugging, set the log output to a file").Short('l').String()

	f *os.File
)

func init() {
	kingpin.Parse()

	if *logout != "" {
		var err error
		f, err = os.Create(*logout)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(f)
	}

	if os.Getenv("KCLI_FAKE") != "true" {
		kafka.Connect(*addrs)
	}
}

func main() {

	var err error
	g, err = ui.NewGui(ui.OutputNormal)
	if err != nil {
		log.Fatal("could not create gui", err)
	}

	w, h := g.Size()
	g.SetManagerFunc(views.GetLayout(w, h))
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

	if err := views.Keybindings(g); err != nil {
		log.Println(err)
		return
	}

	if err := g.MainLoop(); err != nil {
		if err != ui.ErrQuit {
			log.Println(err)
			return
		}
	}

	closed = true
	g.Close()
	if views.After != nil {
		views.After()
	}
}
