package views

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	ui "github.com/jroimartin/gocui"
)

const (
	chars = ` abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890.-,_ +/()*&^%$#@!:"`
	nums  = "1234567890"
)

type footer struct {
	name string

	//kafka or kinesis
	source   string
	coords   coords
	function string
	locked   bool
	width    int
	setView  func(string)

	jump   func(int64, string) error
	offset func(int64) error
	search chan<- string
}

func newFooter(g *ui.Gui, w, h int, source string, ch <-chan string, jump func(int64, string) error, offset func(int64) error, search chan<- string) *footer {
	f := &footer{
		name:   "footer",
		source: source,
		coords: coords{x1: -1, y1: h - 2, x2: w, y2: h},
		width:  w,
		jump:   jump,
		offset: offset,
		search: search,
	}
	go f.flashMessage(g, ch)
	return f
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
		v.Write([]byte(c1(s)))
		v.SetCursor(len(s), 0)
	} else if f.acceptable(in) {
		fmt.Fprint(v, c1(in))
		s = v.Buffer()
		v.SetCursor(len(s)-1, 0)
	}
}

func (f *footer) enter(g *ui.Gui, function string) error {
	v, err := g.View("footer")
	if err != nil {
		return err
	}

	v.Clear()
	v.Write([]byte(c1(fmt.Sprintf("%s: ", function))))
	f.function = function
	return v.SetCursor(len(function)+2, 0)
}

func (f *footer) bail(g *ui.Gui, v *ui.View) error {
	f.setView("body")
	v.Clear()
	return nil
}

func (f *footer) exit(g *ui.Gui, v *ui.View) error {
	s := v.Buffer()
	i := strings.Index(s, ":")
	if i == -1 {
		return nil
	}
	term := strings.TrimSpace(s[i+1:])
	switch f.function {
	case "jump":
		var n int64
		var s string
		if f.source == "kafka" {
			var err error
			n, err = strconv.ParseInt(strings.TrimSpace(term), 10, 64)
			if err != nil {
				return err
			}
		} else {
			d, err := time.ParseDuration(term)
			if err != nil {
				return err
			}
			s = "s"
			n = int64(d.Seconds())
		}
		if err := f.jump(n, s); err != nil {
			return err
		}
	case "search":
		v.Clear()
		f.search <- term
		return f.bail(g, v)
	case "offset":
		n, err := strconv.ParseInt(strings.TrimSpace(term), 10, 64)
		if err != nil {
			return err
		}
		if err := f.offset(n); err != nil {
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
		return f.isDur(s)
	case "offset":
		return f.isNum(s)
	default:
		return false
	}
}

func (f *footer) isChar(s string) bool {
	return strings.Contains(chars, s)
}

func (f *footer) isNum(s string) bool {
	x := nums
	if f.function == "offset" {
		x += "-"
	}
	return strings.Contains(x, s)
}

func (f *footer) isDur(s string) bool {
	x := nums
	if f.source == "kafka" && f.function == "offset" {
		x += "-"
	} else if f.source == "kinesis" {
		x += "smh"
	}
	return strings.Contains(x, s)
}

func (f *footer) flashMessage(g *ui.Gui, ch <-chan string) {
	dur := time.Second * 100000
	for {
		select {
		case m := <-ch:
			dur = time.Second * 3
			f.writeMsg(g, m)
		case <-time.After(dur):
			dur = time.Second * 100000
			f.writeMsg(g, "")
		}
	}
}

func (f *footer) writeMsg(g *ui.Gui, msg string) {
	// if foot.locked {
	// 	return
	// }

	g.Update(func(g *ui.Gui) error {
		v, _ := g.View("footer")
		v.Clear()
		fmt.Fprint(v, msg)
		return nil
	})
}
