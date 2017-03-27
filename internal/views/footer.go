package views

type footer struct {
	name   string
	coords coords
}

func newFooter(w, h int) *footer {
	return &footer{
		name:   "footer",
		coords: coords{x1: -1, y1: h - 2, x2: w, y2: h},
	}
}
