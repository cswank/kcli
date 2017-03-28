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
	fake   = kingpin.Flag("fake", "don't connect to kafka, use hard coded fake data instead").Short('f').Bool()
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

	if !*fake {
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

	if err := views.Keybindings(g); err != nil {
		g.Close()
		log.Fatal(err)
	}

	if err := g.MainLoop(); err != nil {
		if err != ui.ErrQuit {
			g.Close()
			log.Fatal(err)
		}
	}

	if f != nil {
		f.Close()
	}
	g.Close()
}
