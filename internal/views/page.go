package views

import "github.com/cswank/kcli/internal/kafka"

type page struct {
	page   int
	header string
	body   [][]kafka.Row
	cursor int

	next    func(int, string) (page, error)
	forward func(int, string) ([]kafka.Row, error)
	back    func(int, string) ([]kafka.Row, error)
}

type pages struct {
	p []page
}

func (p *pages) header() string {
	l := len(p.p)
	return p.p[l-1].header
}

func (p *pages) cursor() int {
	l := len(p.p)
	return p.p[l-1].cursor
}

func (p *pages) body(page int) []kafka.Row {
	l := len(p.p)
	return p.p[l-1].body[page]
}

func (p *pages) sel(cur int) (page, kafka.Row) {
	l := len(p.p)
	page := p.p[l-1]
	page.cursor = cur
	p.p[l-1] = page
	return page, page.body[page.page][cur]
}

func (p *pages) pop() {
	l := len(p.p)
	if l == 1 {
		return
	}
	p.p = p.p[:l-1]
}

func (p *pages) add(n page) {
	p.p = append(p.p, n)
}
