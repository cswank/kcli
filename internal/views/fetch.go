package views

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/cswank/kcli/internal/colors"
	"github.com/cswank/kcli/internal/kafka"
)

var (
	errNoContent = errors.New("no content")
)

//getTopics -> getTopic -> getPartition -> getMessage
func getTopics(size int, _ interface{}) (page, error) {
	topics, err := kafka.GetTopics()
	if err != nil {
		return page{}, err
	}

	sort.Strings(topics)

	return page{
		name:   "topics",
		header: "topics",
		body:   getTopicsRows(size, topics),
		next:   getTopic,
	}, nil
}

func getTopicsRows(size int, topics []string) [][]row {
	r := make([]row, len(topics))
	for i, t := range topics {
		r[i] = row{args: t, value: t}
	}
	return split(r, size)
}

func getTopic(size int, i interface{}) (page, error) {
	topic, ok := i.(string)
	if !ok {
		return page{}, fmt.Errorf("getTopic could not accept arg: %v", i)
	}

	partitions, err := kafka.GetTopic(topic)
	if err != nil {
		return page{}, err
	}

	return page{
		name:   "topic",
		header: c1("partition     1st offset             current offset         last offset            size"),
		body:   getTopicRows(size, partitions),
		next:   getPartition,
	}, nil
}

func getTopicRows(size int, partitions []kafka.Partition) [][]row {
	r := make([]row, len(partitions))
	tpl := colors.Green("%-13d %-22d %-22d %-22d %d")
	for i, p := range partitions {
		r[i] = row{args: p, value: fmt.Sprintf(tpl, p.Partition, p.Start, p.Offset, p.End, p.End-p.Start), truncate: true}
	}
	return split(r, size)
}

func getPartition(size int, i interface{}) (page, error) {
	partition, ok := i.(kafka.Partition)
	if !ok {
		return page{}, fmt.Errorf("getPartition could not accept arg: %v", i)
	}

	if partition.End == 0 {
		return page{}, errNoContent
	}

	var f func([]byte) bool
	if partition.Filter != "" {
		f = func(val []byte) bool {
			return bytes.Contains(val, []byte(partition.Filter))
		}
	} else {
		f = func(val []byte) bool { return true }
	}

	rows, err := getPartitionRows(size, partition, f)
	if err != nil {
		return page{}, err
	}

	return page{
		name:    "partition",
		header:  c1(fmt.Sprintf("offset       message    topic: %s partition: %d start: %d end: %d", partition.Topic, partition.Partition, partition.Start, partition.End)),
		body:    split(rows, size),
		next:    getMessage,
		forward: nextPartitionPage,
	}, nil
}

func nextPartitionPage() ([]row, error) {
	page := pg.current()
	r := page.lastRow()
	msg, ok := r.args.(kafka.Msg)
	if !ok {
		return nil, fmt.Errorf("wrong arg: %v", r.args)
	}

	p := msg.Partition
	if p.Offset+1 >= p.End {
		return nil, nil
	}
	p.Offset = msg.Offset + 1

	var f func([]byte) bool
	if page.filter && page.search != "" {
		f = func(val []byte) bool {
			return bytes.Contains(val, []byte(page.search))
		}
	} else {
		f = func(val []byte) bool { return true }
	}

	return getPartitionRows(bod.size, p, f)
}

func getPartitionRows(size int, partition kafka.Partition, f func([]byte) bool) ([]row, error) {
	msgs, err := kafka.GetPartition(partition, size, f)
	if err != nil {
		return nil, err
	}

	return getMsgsRows(msgs), nil
}

func getMsgsRows(msgs []kafka.Msg) []row {
	r := make([]row, len(msgs))
	for i, m := range msgs {
		r[i] = row{
			truncate: true,
			args:     m,
			value:    fmt.Sprintf("%-12d %s", m.Partition.Offset, string(m.Value)),
		}
	}

	return r
}

func getMessage(size int, i interface{}) (page, error) {
	msg, ok := i.(kafka.Msg)
	if !ok {
		return page{}, fmt.Errorf("getMessage could not accept arg: %v", i)
	}

	buf, err := getPrettyMsg(msg.Value)
	if err != nil {
		return page{}, err
	}

	var out []row
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		out = append(out, row{value: scanner.Text(), truncate: false})
	}

	return page{
		name:   "message",
		header: c1(fmt.Sprintf("topic: %s partition: %d offset: %d", msg.Partition.Topic, msg.Partition.Partition, msg.Offset)),
		body:   split(out, size),
	}, nil
}

func getPrettyMsg(data []byte) (io.Reader, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		//not json, so return original data
		return bytes.NewBuffer(data), nil
	}

	d, err := colors.Marshal(m)
	buf := bytes.NewBuffer(d)
	return buf, err
}

func split(rows []row, size int) [][]row {
	var out [][]row
	for len(rows) > 0 {
		end := size
		if size > len(rows) {
			end = len(rows)
		}
		out = append(out, rows[:end])
		rows = rows[end:]
	}
	return out
}
