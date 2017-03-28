package views

import (
	"log"

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
)

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

	currentView = bod.name

	p, err := getTopics(bod.size, "")
	if err != nil {
		log.Fatal(err)
	}

	pg = pages{
		p: []page{p},
	}

	return func(g *ui.Gui) error {
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

		_, err = g.SetCurrentView(currentView)
		return err
	}
}

//getTopics -> getTopic -> getPartition -> getMessage
func getTopics(size int, args string) (page, error) {
	r, err := kafka.GetTopics(size, args)
	if err != nil {
		return page{}, err
	}

	return page{
		name:   "topics",
		header: "topics",
		body:   r,
		next:   getTopic,
	}, nil
}

func getTopic(s int, args string) (page, error) {
	r, err := kafka.GetTopic(s, args)
	if err != nil {
		return page{}, err
	}

	return page{
		name:   "topic",
		header: args,
		body:   r,
		next:   getPartition,
	}, nil
}

func getPartition(s int, args string) (page, error) {
	r, err := kafka.GetPartition(s, args)
	if err != nil {
		return page{}, err
	}

	return page{
		name:   "partition",
		header: args,
		body:   r,
		next:   getMessage,
	}, nil
}

func getMessage(s int, args string) (page, error) {
	r, err := kafka.GetMessage(s, args)
	if err != nil {
		return page{}, err
	}

	return page{
		name:   "message",
		header: args,
		body:   r,
	}, nil
}

func next(g *ui.Gui, v *ui.View) error {
	_, cur := v.Cursor()
	if cur < bod.size-1 {
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

//sel gets called when the user hits the enter key.
//The item under the cursor is selected and the next()
//func is called to get then next page.
func sel(g *ui.Gui, v *ui.View) error {
	_, cur := v.Cursor()
	_, size := v.Size()

	p, r := pg.sel(cur)
	n, err := p.next(size, r.Args)
	if err != nil {
		return err
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
	return nil
}

func quit(g *ui.Gui, v *ui.View) error {
	return ui.ErrQuit
}
