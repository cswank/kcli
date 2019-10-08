package views

import ui "github.com/jroimartin/gocui"

type keyHelp struct {
	key  string
	body string
}

type key struct {
	views      []string
	keys       []binding
	keybinding func(*ui.Gui, *ui.View) error

	help struct {
		key  string
		body string
	}
}

type binding interface{}

func (s *screen) getKeys() []key {
	return []key{
		{views: []string{s.body.name}, keys: []binding{'n', ui.KeyArrowDown}, keybinding: s.locked(s.body.next), help: keyHelp{key: "n", body: "(or down arrow) move cursor down"}},
		{views: []string{s.body.name}, keys: []binding{'p', ui.KeyArrowUp}, keybinding: s.locked(s.body.prev), help: keyHelp{key: "p", body: "(or up arrow) move cursor up"}},
		{views: []string{s.body.name}, keys: []binding{'f', ui.KeyArrowRight}, keybinding: s.locked(s.body.forward), help: keyHelp{key: "f", body: "(or right arrow) forward to next page"}},
		{views: []string{s.body.name}, keys: []binding{'b', ui.KeyArrowLeft}, keybinding: s.locked(s.body.back), help: keyHelp{key: "b", body: "(or left arrow) backward to prev page"}},
		{views: []string{s.body.name}, keys: []binding{ui.KeyEnter}, keybinding: s.locked(s.enter), help: keyHelp{key: "enter", body: "view item at cursor"}},
		{views: []string{s.body.name}, keys: []binding{ui.KeyEsc}, keybinding: s.locked(s.escape), help: keyHelp{key: "esc", body: "back to previous view"}},
		{views: []string{s.body.name}, keys: []binding{ui.KeyCtrlJ}, keybinding: s.locked(s.jump), help: keyHelp{key: "C-j", body: "jump to a kafka offset"}},
		{views: []string{s.body.name}, keys: []binding{ui.KeyCtrlO}, keybinding: s.locked(s.offset), help: keyHelp{key: "C-o", body: "set the offset in all partitions of topic"}},
		{views: []string{s.body.name}, keys: []binding{ui.KeyCtrlS, '/'}, keybinding: s.locked(s.search), help: keyHelp{key: "C-s", body: "(or /) search kafka messages"}},
		{views: []string{s.body.name}, keys: []binding{ui.KeyCtrlP}, keybinding: s.locked(s.dump), help: keyHelp{key: "C-p", body: "quit and print to stdout"}},
		{views: []string{s.body.name}, keys: []binding{ui.KeyCtrlD, ui.KeyCtrlC}, keybinding: s.quit, help: keyHelp{key: "C-d (or C-c)", body: "quit"}},
		{views: []string{s.footer.name}, keys: []binding{ui.KeyEnter}, keybinding: s.footer.exit},
		{views: []string{s.footer.name}, keys: []binding{ui.KeyEsc}, keybinding: s.footer.bail},
		{views: []string{s.body.name}, keys: []binding{'h'}, keybinding: s.showHelp, help: keyHelp{key: "h", body: "toggle help"}},
		{views: []string{s.help.name}, keys: []binding{'h'}, keybinding: s.hideHelp},
	}
}
