package views

import (
	"bytes"
	"fmt"
	"log"
	"time"

	ui "github.com/jroimartin/gocui"
)

var (
	helpWidth  = 47
	helpHeight = 15
	tpl        = `%s             C-x means Control x`
	helpMsg    []byte
)

type help struct {
	name   string
	coords coords
	esc    chan bool
	movie  bool
}

func newHelp(w, h int) *help {
	return &help{
		name: "help",
		esc:  make(chan bool),
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

func (h *help) jump(g *ui.Gui, v *ui.View) error {
	v.Clear()
	h.movie = true
	go h.doJump(g, v)
	return nil
}

func (h *help) doJump(g *ui.Gui, v *ui.View) {
	w := helpWidth
	helpWidth = 90
	coords := getHelpCoords(g)
	var err error
	v, err = g.SetView("movie", coords.x1, coords.y1, coords.x2, coords.y2)
	if err != ui.ErrUnknownView {
		log.Println("coudn't create movie view")
		return
	}

	v.SetCursor(0, 0)

	tpl := "    %s\n"
	for _, step := range jumpSteps {
		g.Execute(func(g *ui.Gui) error {
			v.SetCursor(0, 0)
			v.Clear()
			fmt.Fprintf(v, tpl, c1(step.caption))
			fmt.Fprintf(v, step.template, step.args...)
			return nil
		})
		select {
		case <-h.esc:
			currentView = bod.name
			break
		case <-time.After(step.duration):
		}
	}

	helpWidth = w
	v.Clear()
	g.DeleteView("movie")
	g.SetCurrentView(hlp.name)
	currentView = hlp.name
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
	if h.movie {
		h.esc <- true
	}
	return err
}

func getHelpMsg() []byte {
	out := &bytes.Buffer{}
	for _, key := range keys {
		h := key.help
		if h.key != "" {
			fmt.Fprintf(out, fmt.Sprintf("%s %s\n", c3(h.key), c1(fmt.Sprintf(fmt.Sprintf("%%%ds", helpWidth-len(h.key)-4), h.body))))
		}
	}
	return []byte(fmt.Sprintf(tpl, out.String()))
}
