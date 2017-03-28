package views

import (
	"fmt"

	"github.com/cswank/kcli/internal/colors"
	ui "github.com/jroimartin/gocui"
)

var (
	tpl = `  %s %s
  %s %s
  %s %s
  %s %s
  %s %s
  %s %s
  %s %s
  %s %s
  %s %s
  %s %s`

	helpStr = []byte(fmt.Sprintf(
		tpl,
		colors.Yellow("n"),
		colors.White("(or down arrow) move cursor down"),
		colors.Yellow("p"),
		colors.White("(or up arrow) move cursor up"),
		colors.Yellow("f"),
		colors.White("(or right arrow) forward to next page"),
		colors.Yellow("b"),
		colors.White("(or left arrow) backward to prev page"),
		colors.Yellow("enter"),
		colors.White("view item at cursor"),
		colors.Yellow("esc"),
		colors.White("back to previous view"),
		colors.Yellow("j"),
		colors.White("jump to a kafka offset"),
		colors.Yellow("c"),
		colors.White("copy item at cursor to clipboard"),
		colors.Yellow("h"),
		colors.White("toggle help"),
		colors.Yellow("q"),
		colors.White("quit"),
	))
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
	width := 44
	height := 11
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

	v.Write([]byte(helpStr))
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
