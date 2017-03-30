package views

import "github.com/cswank/kcli/internal/kafka"

type page struct {
	name   string
	page   int
	header string
	body   [][]row
	cursor int

	next    func(int, interface{}) (page, error)
	forward func() ([]row, error)
	back    func() error
}

func (p *page) lastRow() row {
	r := p.body[p.page]
	return r[len(r)-1]
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

func (p *pages) body() []row {
	l := len(p.p)
	page := p.p[l-1]
	return page.body[page.page]
}

func (p *pages) current() page {
	l := len(p.p)
	return p.p[l-1]
}

func (p *pages) currentRow() row {
	l := len(p.p)
	page := p.p[l-1]
	return page.body[page.page][page.cursor]
}

func (p *pages) sel(cur int) (page, row) {
	l := len(p.p)
	page := p.p[l-1]
	page.cursor = cur
	p.p[l-1] = page
	return page, page.body[page.page][cur]
}

func (p *pages) pop() page {
	l := len(p.p)
	if l == 1 {
		return page{}
	}

	out := p.p[l-1]
	p.p = p.p[:l-1]
	return out
}

func (p *pages) add(n page) {
	p.p = append(p.p, n)
}

func (p *pages) search(s string) error {
	row := p.currentRow()
	msg := row.args.(kafka.Msg)
	n, err := kafka.Search(msg.Partition, s)
	if err != nil || n == int64(-1) {
		return err
	}

	return p.jump(n)
}

func (p *pages) jump(n int64) error {
	page := pg.pop()
	row := page.body[0][0]
	msg := row.args.(kafka.Msg)
	part := msg.Partition
	part.Offset = n
	var err error
	page, err = getPartition(bod.size, part)
	if err != nil {
		return err
	}

	p.add(page)
	return nil
}

func (p *pages) forward() error {
	l := len(p.p)
	if l == 0 {
		return nil
	}

	page := p.p[l-1]
	if page.page < len(page.body)-1 {
		page.page++
		p.p[l-1] = page
		return nil
	}

	if page.forward == nil {
		return nil
	}

	rows, err := page.forward()
	if err != nil {
		return err
	}

	page.body = append(page.body, rows)
	page.page++
	p.p[l-1] = page
	return nil
}

func (p *pages) back() error {
	l := len(p.p)
	if l == 0 {
		return nil
	}

	page := p.p[l-1]
	if page.page == 0 {
		return nil
	}

	page.page--
	p.p[l-1] = page
	return nil
}
