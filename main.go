package main

import (
	"log"

	"github.com/cswank/kcli/internal/views"

	ui "github.com/jroimartin/gocui"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	g     *ui.Gui
	addrs = kingpin.Flag("addresses", "comma seperated list of kafka addresses").Default("localhost:9092").Short('a').Strings()
)

func main() {
	kingpin.Parse()

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

	g.Close()
}
