package views

import (
	ui "github.com/jroimartin/gocui"
)

type keyHelp struct {
	key  string
	body string
}

type key struct {
	name string
	keys []interface{}
	f    func(*ui.Gui, *ui.View) error

	help struct {
		key  string
		body string
	}
}

var (
	keys []key
)

func getKeys() []key {
	return []key{
		{name: bod.name, keys: []interface{}{ui.KeyCtrlN, ui.KeyArrowDown}, f: next, help: keyHelp{key: "C-n", body: "(or down arrow) move cursor down"}},
		{name: bod.name, keys: []interface{}{ui.KeyCtrlP, ui.KeyArrowUp}, f: prev, help: keyHelp{key: "C-p", body: "(or up arrow) move cursor up"}},
		{name: bod.name, keys: []interface{}{ui.KeyCtrlF, ui.KeyArrowRight}, f: forward, help: keyHelp{key: "C-f", body: "(or right arrow) forward to next page"}},
		{name: bod.name, keys: []interface{}{ui.KeyCtrlB, ui.KeyArrowLeft}, f: back, help: keyHelp{key: "C-b", body: "(or left arrow) backward to prev page"}},
		{name: bod.name, keys: []interface{}{ui.KeyEnter}, f: sel, help: keyHelp{key: "enter", body: "view item at cursor"}},
		{name: bod.name, keys: []interface{}{ui.KeyEsc}, f: popPage, help: keyHelp{key: "esc", body: "back to previous view"}},
		{name: bod.name, keys: []interface{}{'d'}, f: dump, help: keyHelp{key: "d", body: "dump to stdout"}},
		{name: bod.name, keys: []interface{}{'j'}, f: jump, help: keyHelp{key: "j", body: "jump to a kafka offset"}},
		{name: bod.name, keys: []interface{}{'s', '/'}, f: search, help: keyHelp{key: "s", body: "(or /) search kafka messages"}},
		{name: bod.name, keys: []interface{}{'f'}, f: filter, help: keyHelp{key: "f", body: "filter kafka messages"}},
		{name: bod.name, keys: []interface{}{'F'}, f: clearFilter, help: keyHelp{key: "F", body: "clear filter"}},
		{name: foot.name, keys: []interface{}{ui.KeyEnter}, f: foot.exit},
		{name: foot.name, keys: []interface{}{ui.KeyEsc}, f: foot.bail},
		{name: bod.name, keys: []interface{}{'h'}, f: hlp.show, help: keyHelp{key: "h", body: "toggle help"}},
		{name: hlp.name, keys: []interface{}{'h'}, f: hlp.hide},
		{name: bod.name, keys: []interface{}{'q', ui.KeyCtrlC}, f: quit, help: keyHelp{key: "q", body: "(or C-c) quit"}},
		{name: hlp.name, keys: []interface{}{'q', ui.KeyCtrlC}, f: quit},
	}
}

func Keybindings(g *ui.Gui) error {

	for _, k := range keys {
		for _, kb := range k.keys {
			if err := g.SetKeybinding(k.name, kb, ui.ModNone, k.f); err != nil {
				return err
			}
		}
	}
	return nil
}
