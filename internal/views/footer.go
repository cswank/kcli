package views

import (
	"fmt"
	"strconv"
	"strings"

	ui "github.com/jroimartin/gocui"
)

var (
	chars = ` abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890.-,_ +/()*&^%$#@!:"`
	nums  = "1234567890"
)

type footer struct {
	name     string
	coords   coords
	function string
	locked   bool
}

func newFooter(w, h int) *footer {
	return &footer{
		name:   "footer",
		coords: coords{x1: -1, y1: h - 2, x2: w, y2: h},
	}
}

func (f *footer) resize(w, h int) {
	f.coords = coords{x1: -1, y1: h - 2, x2: w, y2: h}
}

func (f *footer) Edit(v *ui.View, key ui.Key, ch rune, mod ui.Modifier) {
	in := string(ch)
	if key == ui.KeySpace {
		in = " "
	}
	s := strings.TrimSpace(v.Buffer())
	if key == 127 && len(s) > 0+len(f.function)+2 {
		v.Clear()
		s = s[:len(s)-1]
		v.Write([]byte(s))
		v.SetCursor(len(s), 0)
	} else if f.acceptable(in) {
		fmt.Fprint(v, in)
		s = v.Buffer()
		v.SetCursor(len(s)-1, 0)
	}
}

func (f *footer) bail(g *ui.Gui, v *ui.View) error {
	currentView = bod.name
	v.Clear()

	var err error
	v, err = g.SetCurrentView(bod.name)
	if err != nil {
		return err
	}
	return v.SetCursor(0, 0)
}

func (f *footer) exit(g *ui.Gui, v *ui.View) error {
	defer func() {
		f.locked = false
	}()

	s := v.Buffer()
	i := strings.Index(s, ":")
	if i == -1 {
		return nil
	}
	term := s[i+1:]
	switch f.function {
	case "jump":
		n, err := strconv.ParseInt(strings.TrimSpace(term), 10, 64)
		if err != nil {
			return err
		}
		page := pg.pop()
		page.search = ""
		page.filter = false
		pg.add(page)

		if err := pg.jump(n, ""); err != nil {
			return err
		}
	case "search":
		keyLock = true
		v.Clear()
		v.Write([]byte("searching..."))
		searchTrigger <- strings.TrimSpace(term)
		currentView = bod.name
		_, err := g.SetCurrentView(bod.name)
		return err
	case "filter":
		if err := pg.filter(strings.TrimSpace(term)); err != nil {
			return err
		}
	}

	return f.bail(g, v)
}

func (f *footer) acceptable(s string) bool {
	switch f.function {
	case "search":
		return f.isChar(s)
	case "filter":
		return f.isChar(s)
	case "jump":
		return f.isNum(s)
	default:
		return false
	}
}

func (f *footer) isChar(s string) bool {
	return strings.Contains(chars, s)
}

func (f *footer) isNum(s string) bool {
	return strings.Contains(nums, s)
}
