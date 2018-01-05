package views

import (
	"fmt"

	ui "github.com/jroimartin/gocui"
)

type header struct {
	text   string
	name   string
	coords coords
	width  int
}

func newHeader(w, h int) *header {
	return &header{
		text:   "topics",
		name:   "header",
		coords: coords{x1: -1, y1: -1, x2: w, y2: 1},
		width:  w,
	}
}

func (h *header) resize(w, _ int) {
	h.width = w
	h.coords = coords{x1: -1, y1: -1, x2: w, y2: 1}
}

func (h *header) Render(g *ui.Gui, v *ui.View) error {
	v.Clear()
	t := fmt.Sprintf("%%s%%%ds", h.width-len(h.text))
	_, err := v.Write([]byte(c1(fmt.Sprintf(t, h.text, "type 'h' for help"))))
	return err
}
