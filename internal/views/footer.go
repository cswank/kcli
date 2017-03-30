package views

import (
	"fmt"
	"strconv"
	"strings"

	ui "github.com/jroimartin/gocui"
)

var (
	chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890.-,_ +/()*&^%$#@!"
	nums  = "1234567890"
)

type footer struct {
	name     string
	coords   coords
	function string
}

func newFooter(w, h int) *footer {
	return &footer{
		name:   "footer",
		coords: coords{x1: -1, y1: h - 2, x2: w, y2: h},
	}
}

func (f *footer) Edit(v *ui.View, key ui.Key, ch rune, mod ui.Modifier) {
	s := strings.TrimSpace(v.Buffer())
	if key == 127 && len(s) > 0 {
		v.Clear()
		s = s[:len(s)-1]
		v.Write([]byte(s))
		v.SetCursor(len(s), 0)
	} else if f.acceptable(string(ch)) {
		fmt.Fprint(v, string(ch))
		s = v.Buffer()
		v.SetCursor(len(s)-1, 0)
	}
}

func (f *footer) exit(g *ui.Gui, v *ui.View) error {
	currentView = bod.name
	s := v.Buffer()
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return nil
	}

	switch f.function {
	case "jump":
		n, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			return err
		}
		if err := pg.jump(n); err != nil {
			return err
		}
	case "search":
		if err := pg.search(strings.TrimSpace(parts[1])); err != nil {
			return err
		}
	}

	v.Clear()
	_, err := g.SetCurrentView(bod.name)
	return err
}

func (f *footer) acceptable(s string) bool {
	switch f.function {
	case "search":
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
