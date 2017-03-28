package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"

	"github.com/Shopify/sarama"
)

var (
	GetTopics, GetTopic, GetPartition, GetMessage fetcher

	addrs []string
	cli   sarama.Client
)

type Row struct {
	Args string
	Data string
}

type fetcher func(int, string) ([][]Row, error)

type PartitionInfo struct {
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
	Start     int64  `json:"start"`
	End       int64  `json:"end"`
	Offset    int64  `json:"cursor"`
	Msg       string `json:"msg"`
}

func init() {
	GetTopics = getMockTopics
	GetTopic = getMockTopic
	GetPartition = getMockPartition
	GetMessage = getMockMessage
}

func (p *PartitionInfo) String() string {
	d, _ := json.Marshal(p)
	return string(d)
}

func Connect(a []string) {
	addrs = a
	var err error
	cli, err = sarama.NewClient(addrs, nil)
	if err != nil {
		log.Fatal(err)
	}

	GetTopics = getTopics
	GetTopic = getTopic
	GetPartition = getPartition
	GetMessage = getMessage
}

func getTopics(size int, args string) ([][]Row, error) {
	topics, err := cli.Topics()
	log.Println("topics", topics, err)
	if err != nil {
		return nil, err
	}

	sort.Strings(topics)
	out := make([]Row, len(topics))

	for i, topic := range topics {
		out[i] = Row{
			Args: topic,
			Data: topic,
		}
	}

	return split(out, size), nil
}

func getTopic(size int, topic string) ([][]Row, error) {
	partitions, err := cli.Partitions(topic)

	if err != nil {
		return nil, err
	}

	out := make([]Row, len(partitions))

	for i, p := range partitions {
		n, err := cli.GetOffset(topic, p, sarama.OffsetNewest)
		if err != nil {
			return nil, err
		}

		o, err := cli.GetOffset(topic, p, sarama.OffsetOldest)
		if err != nil {
			return nil, err
		}

		pi := PartitionInfo{
			Topic:     topic,
			Partition: p,
			Start:     o,
			End:       n,
			Offset:    o,
		}

		d, err := json.Marshal(pi)
		if err != nil {
			return nil, err
		}
		out[i] = Row{Args: string(d), Data: fmt.Sprintf("%d", p)}
	}

	return split(out, size), nil
}

func getPartition(size int, args string) ([][]Row, error) {
	return nil, nil
}

func getMessage(size int, args string) ([][]Row, error) {
	return nil, nil
}

func split(rows []Row, size int) [][]Row {
	var out [][]Row
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
