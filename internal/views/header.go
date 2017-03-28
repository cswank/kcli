package views

import (
	ui "github.com/jroimartin/gocui"
)

type header struct {
	name   string
	coords coords
}

func newHeader(w, h int) *header {
	return &header{
		name:   "header",
		coords: coords{x1: -1, y1: -1, x2: w, y2: 1},
	}
}

func (h *header) Render(g *ui.Gui, v *ui.View) error {
	v.Clear()
	_, err := v.Write([]byte(c1(pg.header())))
	return err
}
