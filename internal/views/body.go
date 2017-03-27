package views

import ui "github.com/jroimartin/gocui"

type body struct {
	size   int
	name   string
	coords coords
	page   int
}

func newBody(w, h int) *body {
	return &body{
		name:   "body",
		size:   h - 2,
		coords: coords{x1: -1, y1: 0, x2: w, y2: h - 1},
	}
}

func (b *body) Render(g *ui.Gui, v *ui.View) error {
	v.Clear()
	body := pg.body(b.page)
	for _, r := range body {
		_, err := v.Write(append([]byte(r.data), []byte("\n")...))
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *body) Select(g *ui.Gui, v *ui.View) error {
	return nil
}
