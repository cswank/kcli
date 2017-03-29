package views

type page struct {
	name   string
	page   int
	header string
	body   [][]row
	cursor int

	next    func(int, interface{}) (page, error)
	forward func(int, interface{}) ([]row, error)
	back    func(int, interface{}) ([]row, error)
}

type pages struct {
	p []page
}

type row struct {
	args  interface{}
	value string
}

func (p *pages) header() string {
	l := len(p.p)
	return p.p[l-1].header
}

func (p *pages) cursor() int {
	l := len(p.p)
	return p.p[l-1].cursor
}

func (p *pages) body(page int) []row {
	l := len(p.p)
	return p.p[l-1].body[page]
}

func (p *pages) current() page {
	l := len(p.p)
	return p.p[l-1]
}

func (p *pages) sel(cur int) (page, row) {
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
