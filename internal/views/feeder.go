package views

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/cswank/kcli/internal/colors"
	"github.com/cswank/kcli/internal/kafka"
)

//feeder feeds the screen the data that it craves
type feeder interface {
	print()
	getRows() ([]string, error)
	page(page int) error
	header() string
	enter(row int) (feeder, error)
	jump(i int64) error
	search(s string, cb func(int64, int64)) (int64, error)
	row() int
}

type root struct {
	cli          *kafka.Client
	width        int
	height       int
	topics       []string
	enteredAt    int
	pg           int
	flashMessage chan<- string
}

func newRoot(cli *kafka.Client, width, height int, flashMessage chan<- string) (*root, error) {
	topics, err := cli.GetTopics()
	if len(topics) == 0 {
		return nil, fmt.Errorf("no topics found in kafka")
	}

	sort.Strings(topics)
	return &root{
		cli:          cli,
		width:        width,
		height:       height,
		topics:       topics,
		flashMessage: flashMessage,
	}, err
}

func (r *root) print() {
	fmt.Println(r.header())
	for _, t := range r.topics {
		fmt.Println(t)
	}
}

func (r *root) page(pg int) error {
	if (r.pg == 0 && pg < 0) || (r.pg+pg)*r.height > len(r.topics) {
		return nil
	}
	r.pg += pg
	return nil
}

func (r *root) getRows() ([]string, error) {
	start := r.pg * r.height
	end := r.pg*r.height + r.height
	if end >= len(r.topics) {
		end = len(r.topics)
	}
	return r.topics[start:end], nil
}

func (r *root) enter(row int) (feeder, error) {
	if row >= len(r.topics) {
		r.flashMessage <- "nothing to see here"
		return nil, errNoData
	}
	r.enteredAt = row
	return newTopic(r.cli, r.topics[row + r.height * r.pg], r.width, r.height, r.flashMessage)
}

func (r *root) jump(_ int64) error { return nil }
func (r *root) search(search string, _ func(int64, int64)) (int64, error) {

	var pa = -1
	var cur = int64(-1)
	for i := range r.topics {
		if r.topics[i] == search {
			pa = i / r.height
			cur = int64(i % r.height)
			r.pg = pa
			break
		}
	}

	r.pg = pa
	return cur, nil
}

func (r *root) row() int { return r.enteredAt }

func (r *root) header() string {
	return "topics"
}

type topic struct {
	cli    *kafka.Client
	height int
	width  int
	offset int

	topic        string
	partitions   []kafka.Partition
	fmt          string
	enteredAt    int
	flashMessage chan<- string
}

func newTopic(cli *kafka.Client, t string, width, height int, flashMessage chan<- string) (feeder, error) {
	partitions, err := cli.GetTopic(t)
	return &topic{
		cli:          cli,
		width:        width,
		height:       height,
		topic:        t,
		partitions:   partitions,
		fmt:          c2("%-13d %-22d %-22d %-22d %d"),
		flashMessage: flashMessage,
	}, err
}

func (t *topic) search(s string, cb func(int64, int64)) (int64, error) {
	results, err := t.cli.SearchTopic(t.partitions, s, false, cb)
	if err != nil || len(results) == 0 {
		return -1, err
	}
	t.partitions = results

	return int64(len(results)), nil
}

func (t *topic) jump(i int64) error {
	if int(i) >= len(t.partitions) || int(i) < 0 {
		t.flashMessage <- "nothing to see here"
		return nil
	}
	t.offset = int(i)
	return nil
}

func (t *topic) row() int { return t.enteredAt }

func (t *topic) header() string {
	return "partition     1st offset             current offset         last offset            size"
}

func (t *topic) setOffset(n int64) error {
	for i, part := range t.partitions {
		if n > 0 {
			end := part.Offset + n
			if end >= part.End {
				end = part.End - 1
				if end < 0 {
					end = 0
				}
			}
			part.Offset = end
		} else {
			end := part.End + n
			if end <= part.Start {
				end = part.Start
			}
			part.Offset = end
		}
		t.partitions[i] = part
	}
	return nil
}

func (t *topic) page(pg int) error {
	offset := t.offset + (t.height * pg)
	if offset > len(t.partitions) {
		return nil
	}
	if offset < 0 {
		offset = 0
	}
	t.offset = offset
	return nil
}

func (t *topic) getRows() ([]string, error) {
	end := t.offset + t.height
	if end >= len(t.partitions) {
		end = len(t.partitions)
	}

	chunk := t.partitions[t.offset:end]
	out := make([]string, len(chunk))
	for i, p := range chunk {
		out[i] = fmt.Sprintf(t.fmt, p.Partition, p.Start, p.Offset, p.End, p.End-p.Start)
	}

	return out, nil
}

func (t *topic) enter(row int) (feeder, error) {
	t.enteredAt = row
	row = t.offset + row
	if row >= len(t.partitions) {
		go func() { t.flashMessage <- "nothing to see here" }()
		return nil, errNoData
	}
	p := t.partitions[row]
	if p.End-p.Start == 0 {
		go func() { t.flashMessage <- "nothing to see here" }()
		return nil, errNoData
	}
	return newPartition(t.cli, p, t.width, t.height, t.flashMessage)
}

func (t *topic) print() {
	fmt.Println(t.header())
	f := t.fmt + "\n"
	for _, p := range t.partitions {
		fmt.Printf(f, p.Partition, p.Start, p.Offset, p.End, p.End-p.Start)
	}
}

type partition struct {
	cli          *kafka.Client
	height       int
	width        int
	partition    kafka.Partition
	rows         []kafka.Message
	enteredAt    int
	fmt          string
	pg           int
	flashMessage chan<- string
}

func newPartition(cli *kafka.Client, p kafka.Partition, width, height int, flashMessage chan<- string) (feeder, error) {
	rows, err := cli.GetPartition(p, height, func(_ []byte) bool { return true })
	return &partition{
		cli:          cli,
		width:        width,
		height:       height,
		partition:    p,
		rows:         rows,
		fmt:          "%-12d %s",
		flashMessage: flashMessage,
	}, err
}

func (p *partition) search(s string, cb func(int64, int64)) (int64, error) {
	i, err := p.cli.Search(p.partition, s, cb)
	if err != nil || i == -1 {
		return i, err
	}

	return i, p.jump(i)
}

func (p *partition) jump(i int64) error {
	if i >= p.partition.End {
		return nil
	}

	p.pg = int(i) / p.height
	p.partition.Offset = i
	rows, err := p.cli.GetPartition(p.partition, p.height, func(_ []byte) bool { return true })
	if err != nil {
		return err
	}
	p.rows = rows
	return nil
}

func (p *partition) row() int { return p.enteredAt }

func (p *partition) header() string {
	return fmt.Sprintf(
		"offset       message    topic: %s partition: %d start: %d end: %d",
		p.partition.Topic,
		p.partition.Partition,
		p.partition.Start,
		p.partition.End,
	)
}

func (p *partition) getRows() ([]string, error) {
	out := make([]string, len(p.rows))
	for i, msg := range p.rows {
		end := p.width
		if len(msg.Value) < end {
			end = len(msg.Value)
		}
		out[i] = fmt.Sprintf(p.fmt, p.partition.Offset+int64(i), string(msg.Value[:end]))
	}

	return out, nil
}

func (p *partition) page(pg int) error {
	if p.pg == 0 && pg < 0 && p.partition.Offset == p.partition.Start {
		return nil
	} else if p.pg == 0 && pg < 0 && p.partition.Offset > p.partition.Start {
		pg = 0
	}

	o := int64((p.pg+pg)*p.height) + p.partition.Start
	if o >= p.partition.End {
		return nil
	}
	p.pg += pg
	p.partition.Offset = o
	rows, err := p.cli.GetPartition(p.partition, p.height, func(_ []byte) bool { return true })
	if err != nil {
		return err
	}
	p.rows = rows
	return nil
}

func (p *partition) enter(row int) (feeder, error) {
	if row >= len(p.rows) {
		go func() { p.flashMessage <- "nothing to see here" }()
		return nil, errNoData
	}
	p.enteredAt = row
	return newMessage(p.rows[row], p.width, p.height, p.flashMessage)
}

func (p *partition) print() {
	p.cli.Fetch(p.partition, p.partition.End, func(s string) {
		fmt.Println(s)
	})
}

type message struct {
	height       int
	width        int
	msg          kafka.Message
	enteredAt    int
	body         []string
	pg           int
	flashMessage chan<- string
}

func newMessage(msg kafka.Message, width, height int, flashMessage chan<- string) (feeder, error) {
	buf, err := prettyMessage(msg.Value)
	if err != nil {
		return nil, err
	}

	var body []string
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		body = append(body, scanner.Text())
	}

	return &message{
		width:        width,
		height:       height,
		msg:          msg,
		body:         body,
		flashMessage: flashMessage,
	}, nil
}

func (m *message) print() {
	for _, r := range m.body {
		fmt.Println(r)
	}
}

func (m *message) search(_ string, _ func(int64, int64)) (int64, error) { return -1, nil }

func (m *message) jump(_ int64) error { return nil }

func (m *message) row() int { return m.enteredAt }

func (m *message) header() string {
	return fmt.Sprintf(
		"topic: %s partition: %d offset: %d",
		m.msg.Partition.Topic,
		m.msg.Partition.Partition,
		m.msg.Offset,
	)
}

func (m *message) page(pg int) error {
	if m.pg == 0 && pg < 0 {
		return nil
	}
	if (pg+m.pg)*m.height > len(m.body) {
		return nil
	}
	m.pg += pg
	return nil
}

func (m *message) enter(row int) (feeder, error) {
	m.enteredAt = row
	return nil, errNoData
}

func (m *message) getRows() ([]string, error) {
	start := m.pg * m.height
	end := start + m.height
	if end >= len(m.body) {
		end = len(m.body)
	}
	return m.body[start:end], nil
}

func prettyMessage(val []byte) (io.Reader, error) {
	var i interface{}
	if err := json.Unmarshal(val, &i); err != nil {
		//not json, so return original data
		return bytes.NewBuffer(val), nil
	}

	d, err := colors.Marshal(i)
	buf := bytes.NewBuffer(d)
	return buf, err
}
