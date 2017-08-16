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
)

func getKeys() []key {
	return []key{
		{views: []string{bod.name}, keys: []interface{}{'n', ui.KeyArrowDown}, keybinding: viewLocked(next), help: keyHelp{key: "n", body: "(or down arrow) move cursor down"}},
		{views: []string{bod.name}, keys: []interface{}{'p', ui.KeyArrowUp}, keybinding: viewLocked(prev), help: keyHelp{key: "p", body: "(or up arrow) move cursor up"}},
		{views: []string{bod.name}, keys: []interface{}{'f', ui.KeyArrowRight}, keybinding: viewLocked(forward), help: keyHelp{key: "f", body: "(or right arrow) forward to next page"}},
		{views: []string{bod.name}, keys: []interface{}{'b', ui.KeyArrowLeft}, keybinding: viewLocked(back), help: keyHelp{key: "b", body: "(or left arrow) backward to prev page"}},
		{views: []string{bod.name}, keys: []interface{}{ui.KeyEnter}, keybinding: viewLocked(sel), help: keyHelp{key: "enter", body: "view item at cursor"}},
		{views: []string{bod.name}, keys: []interface{}{ui.KeyEsc}, keybinding: viewLocked(popPage), help: keyHelp{key: "esc", body: "back to previous view"}},
		{views: []string{bod.name}, keys: []interface{}{ui.KeyCtrlP}, keybinding: viewLocked(dump), help: keyHelp{key: "C-p", body: "quit and print to stdout"}},
		{views: []string{bod.name}, keys: []interface{}{ui.KeyCtrlJ}, keybinding: viewLocked(jump), help: keyHelp{key: "C-j", body: "jump to a kafka offset"}},
		{views: []string{bod.name}, keys: []interface{}{ui.KeyCtrlS, '/'}, keybinding: viewLocked(search), help: keyHelp{key: "C-s", body: "(or /) search kafka messages"}},
		{views: []string{bod.name}, keys: []interface{}{ui.KeyCtrlF}, keybinding: viewLocked(filter), help: keyHelp{key: "C-f", body: "filter kafka messages"}},
		{views: []string{bod.name}, keys: []interface{}{'F'}, keybinding: viewLocked(clearFilter), help: keyHelp{key: "F", body: "clear filter"}},
		{views: []string{foot.name}, keys: []interface{}{ui.KeyEnter}, keybinding: viewLocked(foot.exit)},
		{views: []string{foot.name}, keys: []interface{}{ui.KeyEsc}, keybinding: viewLocked(foot.bail)},
		{views: []string{bod.name}, keys: []interface{}{'h'}, keybinding: viewLocked(hlp.show), help: keyHelp{key: "h", body: "toggle help"}},
		{views: []string{hlp.name}, keys: []interface{}{'h'}, keybinding: viewLocked(hlp.hide)},
		{views: []string{bod.name, hlp.name}, keys: []interface{}{ui.KeyCtrlD, ui.KeyCtrlC}, keybinding: viewLocked(quit), help: keyHelp{key: "C-d (or C-c)", body: "quit"}},

		{views: []string{searchD.name}, keys: []interface{}{ui.KeyCtrlS}, keybinding: searchD.firstResult},
		{views: []string{searchD.name}, keys: []interface{}{'s'}, keybinding: searchD.search},
	}
}

func Keybindings(g *ui.Gui) error {
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
