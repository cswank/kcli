package views

import (
	"fmt"

	"github.com/cswank/kcli/internal/kafka"
)

//getTopics -> getTopic -> getPartition -> getMessage
func getTopics(size int, _ interface{}) (page, error) {
	topics, err := kafka.GetTopics()
	if err != nil {
		return page{}, err
	}

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
		header: topic,
		body:   getTopicRows(size, partitions),
		next:   getPartition,
	}, nil
}

func getTopicRows(size int, partitions []kafka.Partition) [][]row {
	r := make([]row, len(partitions))
	for i, p := range partitions {
		r[i] = row{args: p, value: fmt.Sprintf("%d", p.Partition)}
	}
	return split(r, size)
}

func getPartition(size int, i interface{}) (page, error) {
	partition, ok := i.(kafka.Partition)
	if !ok {
		return page{}, fmt.Errorf("getPartition could not accept arg: %v", i)
	}

	msgs, err := kafka.GetPartition(partition, size)
	if err != nil {
		return page{}, err
	}

	return page{
		name:   "partition",
		header: fmt.Sprintf("%d", partition.Partition),
		body:   getMsgsRows(size, msgs),
		next:   getMessage,
	}, nil
}

func getMsgsRows(size int, msgs []kafka.Msg) [][]row {
	r := make([]row, len(msgs))
	for i, m := range msgs {
		r[i] = row{args: m, value: fmt.Sprintf("%d: %s", m.Partition.Offset, string(m.Value))}
	}
	return split(r, size)
}

func getMessage(size int, i interface{}) (page, error) {
	// msg, ok := i.(kafka.Msg)
	// if !ok {
	// 	return page{}, fmt.Errorf("getMessage could not accept arg: %v", i)
	// }

	return page{
		name:   "message",
		header: "msg",
	}, nil
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
