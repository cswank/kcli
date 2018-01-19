package views

import (
	"fmt"
	"os"
	"strings"

	"github.com/cswank/kcli/internal/colors"
	ui "github.com/jroimartin/gocui"
)

var (
	bg         ui.Attribute
	c1, c2, c3 colors.Colorer
)

func init() {
	bg, c1, c2, c3 = getColors()
}

type coords struct {
	x1 int
	x2 int
	y1 int
	y2 int
}

type Screen struct {
	g      *ui.Gui
	view   string
	height int
	width  int
	lock   bool

	header *header
	body   *body
	footer *footer
	help   *help

	keys []key

	searchChan   <-chan string
	flashMessage chan<- string

	//after is used by the C-p command
	After func()
}

func newScreen(g *ui.Gui, width, height int, opts ...func(*stack) error) (*Screen, error) {
	ch := make(chan string)
	searchCh := make(chan string)
	b, err := newBody(width, height, ch, opts...)
	if err != nil {
		return nil, err
	}

	s := &Screen{
		g:            g,
		view:         "body",
		width:        width,
		height:       height,
		header:       newHeader(width, height),
		body:         b,
		footer:       newFooter(g, width, height, ch, b.jump, b.offset, searchCh),
		help:         newHelp(width, height),
		searchChan:   searchCh,
		flashMessage: ch,
	}

	go s.doSearch()
	s.footer.setView = func(v string) { s.view = v }
	s.keys = s.getKeys()
	return s, nil
}

func (s *Screen) GetLayout(g *ui.Gui, width, height int) func(*ui.Gui) error {
	ui.DefaultEditor = s.footer

	return func(g *ui.Gui) error {
		v, err := g.SetView(s.header.name, s.header.coords.x1, s.header.coords.y1, s.header.coords.x2, s.header.coords.y2)
		if err != nil && err != ui.ErrUnknownView {
			return err
		}

		v.Frame = false
		if err := s.header.Render(g, v); err != nil {
			return err
		}

		v, err = g.SetView(s.body.name, s.body.coords.x1, s.body.coords.y1, s.body.coords.x2, s.body.coords.y2)
		if err != nil && err != ui.ErrUnknownView {
			return err
		}

		v.Frame = false

		if err := s.body.Render(g, v); err != nil {
			return err
		}

		v, err = g.SetView(s.footer.name, s.footer.coords.x1, s.footer.coords.y1, s.footer.coords.x2, s.footer.coords.y2)
		if err != nil && err != ui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Editable = true

		_, err = g.SetCurrentView(s.view)
		return err
	}
}

func getColors() (ui.Attribute, colors.Colorer, colors.Colorer, colors.Colorer) {
	bg = colors.GetBackground(os.Getenv("KCLI_COLOR0"))
	c1 := colors.Get(os.Getenv("KCLI_COLOR1"))
	if c1 == nil {
		c1 = colors.White
	}
	c2 := colors.Get(os.Getenv("KCLI_COLOR2"))
	if c2 == nil {
		c2 = colors.Green
	}
	c3 := colors.Get(os.Getenv("KCLI_COLOR3"))
	if c3 == nil {
		c3 = colors.Yellow
	}
	return bg, c1, c2, c3
}

func (s *Screen) locked(f func(g *ui.Gui, v *ui.View) error) func(g *ui.Gui, v *ui.View) error {
	return func(g *ui.Gui, v *ui.View) error {
		if s.lock {
			return nil
		}
		return f(g, v)
	}
}

func (s *Screen) enter(g *ui.Gui, v *ui.View) error {
	h, err := s.body.enter(g, v)
	s.header.text = h
	return err
}

func (s *Screen) escape(g *ui.Gui, v *ui.View) error {
	h, err := s.body.escape(g, v)
	s.header.text = h
	return err
}

func (s *Screen) quit(g *ui.Gui, v *ui.View) error {
	return ui.ErrQuit
}

func (s *Screen) jump(g *ui.Gui, v *ui.View) error {
	s.view = "footer"
	s.footer.enter(g, "jump")
	return nil
}

func (s *Screen) offset(g *ui.Gui, v *ui.View) error {
	_, ok := s.body.stack.top.(*topic)
	if !ok {
		s.flashMessage <- "you can only set the offset in a topic"
		return nil
	}
	s.view = "footer"
	s.footer.enter(g, "offset")
	return nil
}

func (s *Screen) search(g *ui.Gui, v *ui.View) error {
	s.lock = true
	s.view = "footer"
	s.footer.enter(g, "search")
	return nil
}

func (s *Screen) dump(g *ui.Gui, v *ui.View) error {
	s.After = s.body.stack.top.print
	return ui.ErrQuit
}

func (s *Screen) doSearch() {
	for {
		s.view = "body"
		term := <-s.searchChan
		var i int
		n, err := s.body.search(term, func(a, b int64) {
			if i%10 == 0 {
				s.flashMessage <- fmt.Sprintf(strings.Repeat("|", int(int64(s.width)*a/b)))
			}
			i++
		})

		if err != nil {
			s.flashMessage <- fmt.Sprintf("error: %s", err)
			return
		}

		if n > 0 {
			if s.body.stack.name() == "topic" {
				s.flashMessage <- fmt.Sprintf("%d partitions matched %s", n, term)
			} else {
				s.flashMessage <- fmt.Sprintf("found a match at offset %d", n)
			}
		} else {
			s.flashMessage <- fmt.Sprintf("'%s' not found", term)
		}

		s.g.Update(func(g *ui.Gui) error {
			v, _ := s.g.View("body")
			return s.body.Render(g, v)
		})
		s.lock = false
	}
}

func (s *Screen) showHelp(g *ui.Gui, v *ui.View) error {
	s.view = "help"
	return s.help.show(g, v, s.keys)
}

func (s *Screen) hideHelp(g *ui.Gui, v *ui.View) error {
	s.view = "body"
	return s.help.hide(g, v)
}

func (s *Screen) Keybindings(g *ui.Gui) error {
	for _, k := range s.keys {
		for _, view := range k.views {
			for _, kb := range k.keys {
				if err := g.SetKeybinding(view, kb, ui.ModNone, k.keybinding); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
