package views

import (
	"errors"
	"fmt"
	"strings"

	ui "github.com/jroimartin/gocui"
)

var (
	errNoData = errors.New("nothing to see here")
)

type body struct {
	height       int
	width        int
	name         string
	coords       coords
	rows         []string
	stack        stack
	flashMessage chan<- string
	view         *ui.View
}

func newBody(w, h int, flashMessage chan string) (*body, error) {
	r, err := newRoot(w, h-2, flashMessage)

	return &body{
		name:         "body",
		height:       h - 2,
		width:        w,
		coords:       coords{x1: -1, y1: 0, x2: w, y2: h - 1},
		stack:        newStack(r),
		flashMessage: flashMessage,
	}, err
}

func (b *body) Render(g *ui.Gui, v *ui.View) error {
	b.view = v
	var err error
	b.rows, err = b.stack.top.getRows()
	if err != nil {
		return err
	}

	v.Clear()
	for _, r := range b.rows {
		_, err := v.Write(append([]byte(b.color(r, "", false)), []byte("\n")...))
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

func (b *body) escape(g *ui.Gui, v *ui.View) (string, error) {
	b.stack.pop()
	r := b.stack.top.row()
	return b.stack.top.header(), v.SetCursor(0, r)
}

func (b *body) offset(o int64) error {
	t, ok := b.stack.top.(*topic)
	if !ok {
		return nil
	}
	return t.setOffset(o)
}

func (b *body) enter(g *ui.Gui, v *ui.View) (string, error) {
	_, cur := v.Cursor()
	f, err := b.stack.top.enter(cur)
	if err == errNoData {
		return b.stack.top.header(), nil
	}

	if err != nil {
		return "", err
	}

	b.stack.add(f)
	return b.stack.top.header(), v.SetCursor(0, 0)
}

func (b *body) next(g *ui.Gui, v *ui.View) error {
	_, cur := v.Cursor()
	if cur < len(b.rows)-1 {
		cur++
	}
	return v.SetCursor(0, cur)
}

func (b *body) prev(g *ui.Gui, v *ui.View) error {
	_, cur := v.Cursor()
	if cur > 0 {
		cur--
	}
	return v.SetCursor(0, cur)
}

func (b *body) forward(g *ui.Gui, v *ui.View) error {
	if err := b.stack.top.page(1); err != nil {
		return err
	}
	return v.SetCursor(0, 0)
}

func (b *body) back(g *ui.Gui, v *ui.View) error {
	if err := b.stack.top.page(-1); err != nil {
		return err
	}
	return v.SetCursor(0, 0)
}

func (b *body) jump(i int64) error {
	if err := b.view.SetCursor(0, 0); err != nil {
		return err
	}
	return b.stack.top.jump(i)
}

func (b *body) search(s string, cb func(int64, int64)) (int64, error) {
	return b.stack.top.search(s, cb)
}

type stack struct {
	top     feeder
	feeders []feeder
}

func newStack(f feeder) stack {
	return stack{top: f, feeders: []feeder{f}}
}

func (s *stack) add(f feeder) {
	s.top = f
	s.feeders = append(s.feeders, f)
}

func (s *stack) pop() {
	if len(s.feeders) == 1 {
		return
	}

	i := len(s.feeders) - 1
	s.feeders = s.feeders[0:i]
	s.top = s.feeders[len(s.feeders)-1]
}
