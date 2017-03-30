package views

import (
	ui "github.com/jroimartin/gocui"
)

type key struct {
	name string
	key  interface{}
	mod  ui.Modifier
	f    func(*ui.Gui, *ui.View) error
}

func Keybindings(g *ui.Gui) error {

	keys := []key{
		{bod.name, 'n', ui.ModNone, next},
		{bod.name, ui.KeyArrowDown, ui.ModNone, next},
		{bod.name, 'p', ui.ModNone, prev},
		{bod.name, ui.KeyArrowUp, ui.ModNone, prev},
		{bod.name, 'f', ui.ModNone, forward},
		{bod.name, ui.KeyArrowRight, ui.ModNone, forward},
		{bod.name, 'b', ui.ModNone, back},
		{bod.name, ui.KeyArrowLeft, ui.ModNone, back},
		{bod.name, ui.KeyEnter, ui.ModNone, sel},
		{bod.name, ui.KeyEsc, ui.ModNone, popPage},
		{bod.name, 'd', ui.ModNone, dump},
		{bod.name, 'j', ui.ModNone, jump},
		{bod.name, 's', ui.ModNone, search},
		{foot.name, ui.KeyEnter, ui.ModNone, foot.exit},
		{bod.name, 'h', ui.ModNone, hlp.show},
		{hlp.name, 'h', ui.ModNone, hlp.hide},
		{"", 'q', ui.ModNone, quit},
		{"", ui.KeyCtrlC, ui.ModNone, quit},
	}

	for _, k := range keys {
		if err := g.SetKeybinding(k.name, k.key, k.mod, k.f); err != nil {
			return err
		}
	}
	return nil
}
