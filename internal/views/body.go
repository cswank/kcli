package views

import (
	"fmt"
	"strings"

	ui "github.com/jroimartin/gocui"
)

type body struct {
	size   int
	width  int
	name   string
	coords coords
}

func newBody(w, h int) *body {
	return &body{
		name:   "body",
		size:   h - 2,
		width:  w,
		coords: coords{x1: -1, y1: 0, x2: w, y2: h - 1},
	}
}

func (b *body) Render(g *ui.Gui, v *ui.View) error {
	v.Clear()
	body, page := pg.body()
	for _, r := range body {
		_, err := v.Write(append([]byte(b.color(r.value, page.search, r.truncate)), []byte("\n")...))
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *body) color(val, search string, truncate bool) string {
	if search == "" || !truncate {
		return c2(val)
	}
	i := strings.Index(val, search)
	if i == -1 {
		return c2(val)
	}

	var s1 string
	if i > 13 {
		s1 = fmt.Sprintf("%s...", val[0:13])
	} else {
		s1 = val[0:13]
	}
	s2 := val[i : i+len(search)]
	s3 := val[i+len(search):]
	return fmt.Sprintf("%s%s%s", c2(s1), c3(s2), c2(s3))
}

func (b *body) resize(w, h int) {
	b.size = h - 2
	b.width = w
	b.coords = coords{x1: -1, y1: 0, x2: w, y2: h - 1}
	pg.resize(b.size)
}

func (b *body) Select(g *ui.Gui, v *ui.View) error {
	return nil
}
