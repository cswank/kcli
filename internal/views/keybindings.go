package views

import (
	ui "github.com/jroimartin/gocui"
)

type keyHelp struct {
	key  string
	body string
}

type key struct {
	views      []string
	keys       []interface{}
	keybinding func(*ui.Gui, *ui.View) error

	help struct {
		key  string
		body string
	}
}

var (
	keys []key
	gui  *ui.Gui
)

func getKeys() []key {
	return []key{
		{views: []string{bod.name}, keys: []interface{}{'n', ui.KeyArrowDown}, keybinding: next, help: keyHelp{key: "n", body: "(or down arrow) move cursor down"}},
		{views: []string{bod.name}, keys: []interface{}{'p', ui.KeyArrowUp}, keybinding: prev, help: keyHelp{key: "p", body: "(or up arrow) move cursor up"}},
		{views: []string{bod.name}, keys: []interface{}{'f', ui.KeyArrowRight}, keybinding: forward, help: keyHelp{key: "f", body: "(or right arrow) forward to next page"}},
		{views: []string{bod.name}, keys: []interface{}{'b', ui.KeyArrowLeft}, keybinding: back, help: keyHelp{key: "b", body: "(or left arrow) backward to prev page"}},
		{views: []string{bod.name}, keys: []interface{}{ui.KeyEnter}, keybinding: sel, help: keyHelp{key: "enter", body: "view item at cursor"}},
		{views: []string{bod.name}, keys: []interface{}{ui.KeyEsc}, keybinding: popPage, help: keyHelp{key: "esc", body: "back to previous view"}},
		{views: []string{bod.name}, keys: []interface{}{ui.KeyCtrlP}, keybinding: dump, help: keyHelp{key: "C-p", body: "quit and print to stdout"}},
		{views: []string{bod.name}, keys: []interface{}{ui.KeyCtrlJ}, keybinding: jump, help: keyHelp{key: "C-j", body: "jump to a kafka offset"}},
		{views: []string{bod.name}, keys: []interface{}{ui.KeyCtrlS, '/'}, keybinding: search, help: keyHelp{key: "C-s", body: "(or /) search kafka messages"}},
		{views: []string{bod.name}, keys: []interface{}{ui.KeyCtrlF}, keybinding: filter, help: keyHelp{key: "C-f", body: "filter kafka messages"}},
		{views: []string{bod.name}, keys: []interface{}{'F'}, keybinding: clearFilter, help: keyHelp{key: "F", body: "clear filter"}},
		{views: []string{foot.name}, keys: []interface{}{ui.KeyEnter}, keybinding: foot.exit},
		{views: []string{foot.name}, keys: []interface{}{ui.KeyEsc}, keybinding: foot.bail},
		{views: []string{bod.name}, keys: []interface{}{'h'}, keybinding: hlp.show, help: keyHelp{key: "h", body: "toggle help"}},
		{views: []string{hlp.name}, keys: []interface{}{'h'}, keybinding: hlp.hide},
		{views: []string{bod.name, hlp.name}, keys: []interface{}{ui.KeyCtrlD}, keybinding: quit, help: keyHelp{key: "C-d", body: "quit"}},
	}
}

func Keybindings(g *ui.Gui) error {
	gui = g
	for _, k := range keys {
		for _, view := range k.views {
			for _, kb := range k.keys {
				if err := g.SetKeybinding(view, kb, ui.ModNone, k.keybinding); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
