package views

import (
	"fmt"
	"log"

	"github.com/cswank/kcli/internal/colors"
	"github.com/cswank/kcli/internal/kafka"
	ui "github.com/jroimartin/gocui"
)

var (
	pg pages

	head *header
	bod  *body
	foot *footer
	hlp  *help

	currentView string

	c1, c2, c3 colors.Colorer

	//After gets called by main when the gui is closed (if it's not nil)
	After func()
)

func init() {
	c1, c2, c3 = getColors()
	helpMsg = getHelpMsg()
}

type coords struct {
	x1 int
	x2 int
	y1 int
	y2 int
}

type View interface {
	Render(g *ui.Gui, v *ui.View) error
}

func GetLayout(width, height int) func(g *ui.Gui) error {
	head = newHeader(width, height)
	bod = newBody(width, height)
	foot = newFooter(width, height)
	hlp = newHelp(width, height)

	ui.DefaultEditor = foot

	currentView = bod.name

	p, err := getTopics(bod.size, "")
	if err != nil {
		log.Fatal(err)
	}

	pg = pages{
		p: []page{p},
	}

	return func(g *ui.Gui) error {
		w, h := g.Size()
		if h != height || w != width {
			width = w
			height = h
			head.resize(w, h)
			bod.resize(w, h)
			foot.resize(w, h)
		}
		v, err := g.SetView(head.name, head.coords.x1, head.coords.y1, head.coords.x2, head.coords.y2)
		if err != nil && err != ui.ErrUnknownView {
			return err
		}
		v.Frame = false
		if err := head.Render(g, v); err != nil {
			return err
		}

		v, err = g.SetView(bod.name, bod.coords.x1, bod.coords.y1, bod.coords.x2, bod.coords.y2)
		if err != nil && err != ui.ErrUnknownView {
			return err
		}

		v.Frame = false

		if err := bod.Render(g, v); err != nil {
			return err
		}

		v, err = g.SetView(foot.name, foot.coords.x1, foot.coords.y1, foot.coords.x2, foot.coords.y2)
		if err != nil && err != ui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Editable = true

		_, err = g.SetCurrentView(currentView)
		return err
	}
}

func next(g *ui.Gui, v *ui.View) error {
	_, cur := v.Cursor()
	if cur < bod.size-1 && cur < len(pg.body())-1 {
		cur++
	}
	return v.SetCursor(0, cur)
}

func prev(g *ui.Gui, v *ui.View) error {
	_, cur := v.Cursor()
	if cur > 0 {
		cur--
	}
	return v.SetCursor(0, cur)
}

func forward(g *ui.Gui, v *ui.View) error {
	if err := pg.forward(); err != nil {
		return err
	}
	return v.SetCursor(0, 0)
}

func back(g *ui.Gui, v *ui.View) error {
	if err := pg.back(); err != nil {
		return err
	}
	return v.SetCursor(0, 0)
}

//sel gets called when the user hits the enter key.
//The item under the cursor is selected and the next()
//func is called to get then next page.
func sel(g *ui.Gui, v *ui.View) error {
	_, cur := v.Cursor()
	_, size := v.Size()

	p, r := pg.sel(cur)
	n, err := p.next(size, r.args)
	if err != nil {
		return err
	}

	if len(n.body) == 0 {
		return nil
	}

	pg.add(n)
	return v.SetCursor(0, 0)
}

func popPage(g *ui.Gui, v *ui.View) error {
	pg.pop()
	return v.SetCursor(0, pg.cursor())
}

func jump(g *ui.Gui, v *ui.View) error {
	p := pg.current()
	if p.name != "partition" {
		return nil
	}

	var err error
	currentView = foot.name
	v, err = g.SetCurrentView(foot.name)
	if err != nil {
		return err
	}

	v.Clear()
	v.Write([]byte("jump: "))
	foot.function = "jump"
	return v.SetCursor(6, 0)
}

func search(g *ui.Gui, v *ui.View) error {
	p := pg.current()
	if p.name != "partition" {
		return nil
	}

	var err error
	currentView = foot.name
	v, err = g.SetCurrentView(foot.name)
	if err != nil {
		return err
	}

	v.Clear()
	v.Write([]byte("search: "))
	foot.function = "search"
	return v.SetCursor(8, 0)
}

func dump(g *ui.Gui, v *ui.View) error {
	_, cur := v.Cursor()
	page, r := pg.sel(cur)
	switch page.name {
	case "partition":
		msg := r.args.(kafka.Msg)
		part := msg.Partition
		After = func() {
			kafka.Fetch(part, part.End, func(s string) {
				fmt.Println(s)
			})
		}
	default:
		After = func() {
			fmt.Println(page.header)
			for _, rows := range page.body {
				for _, s := range rows {
					fmt.Println(s.value)
				}
			}
		}
	}

	return ui.ErrQuit
}

func quit(g *ui.Gui, v *ui.View) error {
	return ui.ErrQuit
}
