package views

import (
	ui "github.com/jroimartin/gocui"
)

type help struct {
	name   string
	coords coords
}

func newHelp(w, h int) *help {
	return &help{
		name:   "help",
		coords: getHelpCoords(w, h),
	}
}

func getHelpCoords(maxX, maxY int) coords {
	width := 42
	height := 11
	x1 := maxX/2 - width/2
	x2 := maxX/2 + width/2
	y1 := maxY/2 - height/2
	y2 := maxY/2 + height/2
	return coords{x1: x1, y1: y1, x2: x2, y2: y2}
}

func (h *help) show(g *ui.Gui, v *ui.View) error {
	var err error
	v, err = g.SetView("help", h.coords.x1, h.coords.y1, h.coords.x2, h.coords.y2)
	if err != ui.ErrUnknownView {
		return err
	}

	v.Title = h.name
	v.Write([]byte(c3("help!!!!")))
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
