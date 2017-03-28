package views

import (
	"os"

	"github.com/cswank/kcli/internal/colors"
	ui "github.com/jroimartin/gocui"
)

type body struct {
	size   int
	name   string
	coords coords
	page   int
}

var (
	c1, c2, c3 colors.Colorer
)

func init() {
	c1 = colors.Get(os.Getenv("KCLI_COLOR1"))
	if c1 == nil {
		c1 = colors.White
	}
	c2 = colors.Get(os.Getenv("KCLI_COLOR2"))
	if c2 == nil {
		c2 = colors.Green
	}
	c3 = colors.Get(os.Getenv("KCLI_COLOR3"))
	if c3 == nil {
		c3 = colors.Yellow
	}
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
		_, err := v.Write(append([]byte(c2(r.data)), []byte("\n")...))
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *body) Select(g *ui.Gui, v *ui.View) error {
	return nil
}
