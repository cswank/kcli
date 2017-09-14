package views

import (
	"os"

	"github.com/cswank/kcli/internal/colors"
	ui "github.com/jroimartin/gocui"
)

func getColors() (ui.Attribute, colors.Colorer, colors.Colorer, colors.Colorer) {
	bg = colors.GetBackground(os.Getenv("KCLI_COLOR0"))
	c1 := colors.Get(os.Getenv("KCLI_COLOR1"))
	if c1 == nil {
		c1 = colors.White
	}
	c2 := colors.Get(os.Getenv("KCLI_COLOR2"))
	if c2 == nil {
		c2 = colors.Green
	}
	c3 := colors.Get(os.Getenv("KCLI_COLOR3"))
	if c3 == nil {
		c3 = colors.Yellow
	}
	return bg, c1, c2, c3
}
