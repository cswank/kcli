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
		{bod.name, ui.KeyCtrlN, ui.ModNone, next},
		{bod.name, ui.KeyArrowDown, ui.ModNone, next},
		{bod.name, ui.KeyCtrlP, ui.ModNone, prev},
		{bod.name, ui.KeyArrowUp, ui.ModNone, prev},
		{bod.name, ui.KeyCtrlF, ui.ModNone, forward},
		{bod.name, ui.KeyArrowRight, ui.ModNone, forward},
		{bod.name, ui.KeyCtrlB, ui.ModNone, back},
		{bod.name, ui.KeyArrowLeft, ui.ModNone, back},
		{bod.name, ui.KeyEnter, ui.ModNone, sel},
		{bod.name, ui.KeyEsc, ui.ModNone, popPage},
		{bod.name, 'd', ui.ModNone, dump},
		{bod.name, 'j', ui.ModNone, jump},
		{bod.name, 's', ui.ModNone, search},
		{bod.name, '/', ui.ModNone, search},
		{bod.name, 'f', ui.ModNone, filter},
		{bod.name, 'F', ui.ModNone, clearFilter},
		{foot.name, ui.KeyEnter, ui.ModNone, foot.exit},
		{foot.name, ui.KeyEsc, ui.ModNone, foot.bail},
		{bod.name, 'h', ui.ModNone, hlp.show},
		{hlp.name, 'h', ui.ModNone, hlp.hide},
		{bod.name, 'q', ui.ModNone, quit},
		{hlp.name, 'q', ui.ModNone, quit},
		{bod.name, ui.KeyCtrlC, ui.ModNone, quit},
		{hlp.name, ui.KeyCtrlC, ui.ModNone, quit},
	}

	for _, k := range keys {
		if err := g.SetKeybinding(k.name, k.key, k.mod, k.f); err != nil {
			return err
		}
	}
	return nil
}
