package views

import (
	"fmt"
	"log"

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

type row struct {
	//args are fed to page.next() to get data for the next page
	args string

	//data written to the screen
	data string
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
func getTopics(s int, args string) (page, error) {
	r := make([]row, s)
	for i := 0; i < s; i++ {
		s := fmt.Sprintf("topic %d", i)
		r[i] = row{args: s, data: s}
	}
	return page{
		header: "topics",
		body:   [][]row{r},
		next:   getTopic,
	}, nil
}

func getTopic(s int, args string) (page, error) {
	return page{
		header: args,
		body: [][]row{
			[]row{
				{args: "partition 1", data: "partition 1"},
				{args: "partition 2", data: "partition 2"},
				{args: "partition 3", data: "partition 3"},
			},
		},
		next: getPartition,
	}, nil
}

func getPartition(s int, args string) (page, error) {
	return page{
		header: args,
		body: [][]row{
			[]row{
				{args: `{"offset": 0}`, data: `{"name": "fred"}`},
				{args: `{"offset": 1}`, data: `{"name": "craig"}`},
				{args: `{"offset": 2}`, data: `{"name": "laura"}`},
			},
		},
		next: getMessage,
	}, nil
}

func getMessage(s int, args string) (page, error) {
	return page{
		header: args,
		body: [][]row{
			[]row{
				{data: `{"name": "fred"}`},
			},
		},
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

	p, r := pg.sel(cur)
	n, err := p.next(0, r.args)
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

func quit(g *ui.Gui, v *ui.View) error {
	return ui.ErrQuit
}

type key struct {
	name string
	key  interface{}
	mod  ui.Modifier
	f    func(*ui.Gui, *ui.View) error
}

func Keybindings(g *ui.Gui) error {

	keys := []key{
		{"", ui.KeyCtrlC, ui.ModNone, quit},
		{"", 'q', ui.ModNone, quit},
		{bod.name, 'n', ui.ModNone, next},
		{bod.name, 'p', ui.ModNone, prev},
		{bod.name, ui.KeyEnter, ui.ModNone, sel},
		{bod.name, ui.KeyEsc, ui.ModNone, popPage},
		{bod.name, 'h', ui.ModNone, hlp.show},
		{hlp.name, 'h', ui.ModNone, hlp.hide},
	}

	for _, k := range keys {
		if err := g.SetKeybinding(k.name, k.key, k.mod, k.f); err != nil {
			return err
		}
	}
	return nil
}
