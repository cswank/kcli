package views

import (
	"fmt"

	ui "github.com/jroimartin/gocui"
)

var (
	helpMsg []byte

	tpl = `  %-15s %62s
  %-15s %62s
  %-15s %62s
  %-15s %62s
  %-15s %62s
  %-15s %62s
  %-15s %62s
  %-15s %62s
  %-15s %62s
  %-15s %62s
  %-15s %62s`
)

type help struct {
	name   string
	coords coords
}

func newHelp(w, h int) *help {
	return &help{
		name: "help",
	}
}

func getHelpCoords(g *ui.Gui) coords {
	maxX, maxY := g.Size()
	width := 62
	height := 12
	x1 := maxX/2 - width/2
	x2 := maxX/2 + width/2
	y1 := maxY/2 - height/2
	y2 := maxY/2 + height/2 + height%2
	return coords{x1: x1, y1: y1, x2: x2, y2: y2}
}

func (h *help) show(g *ui.Gui, v *ui.View) error {
	var err error

	coords := getHelpCoords(g)

	v, err = g.SetView("help", coords.x1, coords.y1, coords.x2, coords.y2)
	if err != ui.ErrUnknownView {
		return err
	}

	v.Title = h.name

	v.Write([]byte(helpMsg))
	_, err = g.SetCurrentView("help")
	currentView = h.name
	return err
}

func (h *help) hide(g *ui.Gui, v *ui.View) error {
	v.Clear()
	if err := g.DeleteView(h.name); err != nil {
		return err
	}

	vw, err := g.SetCurrentView(bod.name)
	if err != nil {
		return err
	}

	currentView = bod.name
	vw.Clear()
	return err
}

func getHelpMsg() []byte {
	return []byte(fmt.Sprintf(
		tpl,
		c3("n"),
		c1("(or down arrow) move cursor down"),
		c3("p"),
		c1("(or up arrow) move cursor up"),
		c3("f"),
		c1("(or right arrow) forward to next page"),
		c3("b"),
		c1("(or left arrow) backward to prev page"),
		c3("enter"),
		c1("view item at cursor"),
		c3("esc"),
		c1("back to previous view"),
		c3("j"),
		c1("jump to a kafka offset"),
		c3("d"),
		c1("consume kafka from cursor to end and write to stdout"),
		c3("c"),
		c1("copy item at cursor to clipboard"),
		c3("h"),
		c1("toggle help"),
		c3("q"),
		c1("quit"),
	))
}
