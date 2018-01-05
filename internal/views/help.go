package views

import (
	"bytes"
	"fmt"

	ui "github.com/jroimartin/gocui"
)

var (
	helpWidth  = 49
	helpHeight = 15
	tpl        = `%s             C-x means Control x`
)

type help struct {
	name   string
	coords coords
	body   []byte
}

func newHelp(w, h int) *help {
	return &help{
		name: "help",
	}
}

func getHelpCoords(g *ui.Gui) coords {
	maxX, maxY := g.Size()
	x1 := maxX/2 - helpWidth/2
	x2 := maxX/2 + helpWidth/2
	y1 := maxY/2 - helpHeight/2
	y2 := maxY/2 + helpHeight/2 + helpHeight%2
	return coords{x1: x1, y1: y1, x2: x2, y2: y2}
}

func (h *help) show(g *ui.Gui, v *ui.View, keys []key) error {
	v.Editable = false
	if h.body == nil {
		h.body = h.getBody(keys)
	}

	var err error

	coords := getHelpCoords(g)

	v, err = g.SetView("help", coords.x1, coords.y1, coords.x2, coords.y2)
	if err != ui.ErrUnknownView {
		return err
	}

	v.Title = h.name

	v.Write([]byte(h.body))
	_, err = g.SetCurrentView("help")
	return err
}

func (h *help) hide(g *ui.Gui, v *ui.View) error {
	v.Clear()
	return g.DeleteView(h.name)
}

func (h *help) getBody(keys []key) []byte {
	out := &bytes.Buffer{}
	for _, key := range keys {
		h := key.help
		if h.key != "" {
			fmt.Fprintf(out, fmt.Sprintf("%s %s\n", c3(h.key), c1(fmt.Sprintf(fmt.Sprintf("%%%ds", helpWidth-len(h.key)-4), h.body))))
		}
	}
	return []byte(fmt.Sprintf(tpl, out.String()))
}
