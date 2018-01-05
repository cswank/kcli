package views

import (
	"fmt"

	ui "github.com/jroimartin/gocui"
)

type searchDialog struct {
	name        string
	coords      coords
	term        string
	stopAtFirst bool
	visible     bool
}

func newSearchDialog(w, h int) *searchDialog {
	return &searchDialog{
		name:   "search",
		coords: getSearchCoords(w, h),
	}
}

func getSearchCoords(w, h int) coords {
	return coords{x1: -1, y1: h - 3, x2: w, y2: h}
}

func (s *searchDialog) resize(w, h int) {
	s.coords = getSearchCoords(w, h)
}

func (s *searchDialog) Render(g *ui.Gui, v *ui.View) error {
	c := c1
	if s.stopAtFirst {
		c = c3
	}
	v.Clear()
	v.Write([]byte(c(fmt.Sprintf("C-s: stop searching at first result (%t)\n", s.stopAtFirst))))
	v.Write([]byte(c1(fmt.Sprintf("s:   search for %s\n", s.term))))
	return nil
}

// func (s *searchDialog) show(g *ui.Gui, term string) error {
// 	p := pg.current()
// 	if p.name == "partition" {
// 		currentView = bod.name
// 		_, err := g.SetCurrentView(bod.name)
// 		searchTrigger <- searchItem{term: strings.TrimSpace(term)}
// 		return err
// 	}

// 	s.visible = true
// 	s.term = term
// 	currentView = s.name
// 	return nil
// }

// func (s *searchDialog) firstResult(g *ui.Gui, v *ui.View) error {
// 	s.stopAtFirst = !s.stopAtFirst
// 	return nil
// }

// func (s *searchDialog) search(g *ui.Gui, v *ui.View) error {
// 	s.visible = false
// 	v.Clear()
// 	if err := g.DeleteView(s.name); err != nil {
// 		return err
// 	}

// 	currentView = bod.name
// 	vb, err := g.SetCurrentView(bod.name)
// 	vb.SetCursor(0, 0)
// 	searchTrigger <- searchItem{term: strings.TrimSpace(s.term), firstResult: s.stopAtFirst}
// 	return err
// }
