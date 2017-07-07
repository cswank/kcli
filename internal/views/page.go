package views

import "github.com/cswank/kcli/internal/kafka"

type page struct {
	name   string
	page   int
	header string
	body   [][]row
	cursor int
	search string
	filter bool

	next    func(int, interface{}) (page, error)
	forward func() ([]row, error)
	back    func() error
}

func (p *page) lastRow() row {

	r := p.body[p.page]
	return r[len(r)-1]
}

func (p *page) resize(s int) {
	p.cursor = 0
	p.page = 0
	var rows []row
	for _, r := range p.body {
		rows = append(rows, r...)
	}
	p.body = split(rows, s)
}

type pages struct {
	p []page
}

type row struct {
	args  interface{}
	value string
}

func (p *pages) resize(s int) {
	for i, page := range p.p {
		page.resize(s)
		p.p[i] = page
	}
}

func (p *pages) header() string {
	l := len(p.p)
	return p.p[l-1].header
}

func (p *pages) cursor() int {
	l := len(p.p)
	return p.p[l-1].cursor
}

func (p *pages) body() ([]row, page) {
	l := len(p.p)
	page := p.p[l-1]
	return page.body[page.page], page
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

func (p *pages) lastRow() row {
	l := len(p.p)
	page := p.p[l-1]
	rows := page.body[page.page]
	return rows[len(rows)-1]
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

func (p *pages) filter(s string) error {
	row := p.currentRow()
	msg := row.args.(kafka.Msg)

	if s == "" {
		s = p.current().search
		if s == "" {
			return nil
		}
	}

	page := p.pop()
	page.search = s
	page.filter = true
	p.add(page)
	return p.jump(msg.Offset, s)
}

func (p *pages) search(s string) error {
	r := p.currentRow()
	msg, ok := r.args.(kafka.Msg)
	if ok {
		return p.searchPartition(s, msg)
	}

	return p.searchTopic(s)
}

func (p *pages) searchTopic(s string) error {
	pg := p.current()
	var partitions []kafka.Partition
	var size int
	for i, rows := range pg.body {
		if i == 0 {
			size = len(rows)
		}
		for _, r := range rows {
			partitions = append(partitions, r.args.(kafka.Partition))
		}
	}

	found, err := kafka.SearchTopic(partitions, s)
	if err != nil {
		return err
	}

	if len(found) == 0 {
		msgs <- "not found"
		return nil
	}

	newPg := page{
		name:   "topic",
		header: c1("partition     1st offset             last offset            size"),
		body:   getTopicRows(size, found),
		next:   getPartition,
		search: s,
	}
	p.pop()
	p.add(newPg)
	return nil
}

func (p *pages) searchPartition(s string, msg kafka.Msg) error {
	if s == "" {
		s = p.current().search
		if s == "" {
			return nil
		}
		//using the previous search term, moving one offset
		//forward to skip the previous search result
		msg.Partition.Offset++
	} else {
		page := p.pop()
		page.search = s
		p.add(page)
	}

	n, err := kafka.Search(msg.Partition, s)
	if err != nil || n == int64(-1) {
		msgs <- "not found"
		return err
	}

	return p.jump(n, s)
}

func (p *pages) jump(n int64, s string) error {
	page := pg.pop()
	f := page.filter
	row := page.body[0][0]
	msg := row.args.(kafka.Msg)
	part := msg.Partition
	if page.filter {
		part.Filter = page.search
	}
	if n >= part.End || n < 0 {
		msgs <- "invalid offset"
		p.add(page)
		return nil
	}

	part.Offset = n
	var err error
	page, err = getPartition(bod.size, part)
	if err != nil {
		p.add(page)
		return err
	}

	page.search = s
	page.filter = f
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

	if rows == nil {
		return nil
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

	if page.name == "partition" && page.page == 0 && !page.filter {
		row := page.body[0][0]
		msg := row.args.(kafka.Msg)
		part := msg.Partition
		n := part.Offset - int64(bod.size)
		if n < 0 {
			n = 0
		}
		return p.jump(n, "")
	}

	if page.page == 0 {
		return nil
	}

	page.page--
	p.p[l-1] = page
	return nil
}
