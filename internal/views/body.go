package views

import ui "github.com/jroimartin/gocui"

type body struct {
	size   int
	name   string
	coords coords
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
	body := pg.body()
	for _, r := range body {
		_, err := v.Write(append([]byte(c2(r.value)), []byte("\n")...))
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *body) resize(w, h int) {
	b.size = h - 2
	b.coords = coords{x1: -1, y1: 0, x2: w, y2: h - 1}
	pg.resize(b.size)
}

func (b *body) Select(g *ui.Gui, v *ui.View) error {
	return nil
}
