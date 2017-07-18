package kafka

import (
	"encoding/json"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/Shopify/sarama"
)

var (
	GetTopics func() ([]string, error)
	addrs     []string
	cli       sarama.Client
)

type Partition struct {
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
	Start     int64  `json:"start"`
	End       int64  `json:"end"`
	Offset    int64  `json:"offset"`
	Filter    string `json:"filter"`
}

func (p *Partition) String() string {
	d, _ := json.Marshal(p)
	return string(d)
}

type Msg struct {
	Partition Partition `json:"partition"`
	Value     []byte    `json:"msg"`
	Offset    int64     `json:"offset"`
}

func Connect(a []string) {
	addrs = a
	var err error
	cli, err = sarama.NewClient(addrs, nil)
	if err != nil {
		log.Fatal(err)
	}

	GetTopics = cli.Topics
}

func GetTopic(topic string) ([]Partition, error) {
	partitions, err := cli.Partitions(topic)

	if err != nil {
		return nil, err
	}

	out := make([]Partition, len(partitions))

	for i, p := range partitions {
		n, err := cli.GetOffset(topic, p, sarama.OffsetNewest)
		if err != nil {
			return nil, err
		}

		o, err := cli.GetOffset(topic, p, sarama.OffsetOldest)
		if err != nil {
			return nil, err
		}
		out[i] = Partition{
			Topic:     topic,
			Partition: p,
			Start:     o,
			End:       n,
			Offset:    o,
		}
	}
	return out, nil
}

func GetPartition(part Partition, end int, f func([]byte) bool) ([]Msg, error) {
	c, err := sarama.NewConsumer(addrs, nil)
	if err != nil {
		return nil, err
	}

	pc, err := c.ConsumePartition(part.Topic, part.Partition, part.Offset)
	if err != nil {
		return nil, err
	}

	defer func() {
		c.Close()
		pc.Close()
	}()

	var out []Msg

	var msg *sarama.ConsumerMessage
	var i int
	var last bool
	for i < end && !last {
		select {
		case msg = <-pc.Messages():
			if f(msg.Value) {
				out = append(out, Msg{
					Value:  msg.Value,
					Offset: msg.Offset,
					Partition: Partition{
						Offset:    msg.Offset,
						Partition: msg.Partition,
						Topic:     msg.Topic,
						End:       part.End,
					},
				})
				i++
			}
			last = msg.Offset == part.End-1
		case <-time.After(time.Second):
			break
		}
	}

	return out, nil
}

func Close() {
	if cli != nil {
		cli.Close()
	}
}

type searchResult struct {
	partition Partition
	offset    int64
	error     error
}

func SearchTopic(partitions []Partition, s string, cb func(int, int)) ([]Partition, error) {
	ch := make(chan searchResult)
	n := len(partitions)
	for _, p := range partitions {
		go func(partition Partition, ch chan searchResult) {
			i, err := Search(partition, s)
			ch <- searchResult{partition: partition, offset: i, error: err}
		}(p, ch)
	}

	var results []Partition
	for i := 0; i < len(partitions); i++ {
		r := <-ch
		cb(i, n)
		if r.error != nil {
			return nil, r.error
		}
		if r.offset > -1 {
			r.partition.Offset = r.offset
			results = append(results, r.partition)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[j].Partition >= results[i].Partition
	})

	return results, nil
}

func Search(info Partition, s string) (int64, error) {
	n := int64(-1)
	var i int64
	err := consume(info, info.End, func(msg string) bool {
		if strings.Contains(msg, s) {
			n = i + info.Offset
			return true
		}
		i++
		return false
	})

	return n, err
}

func Fetch(info Partition, end int64, cb func(string)) error {
	return consume(info, end, func(s string) bool {
		cb(s)
		return false
	})
}

func consume(info Partition, end int64, cb func(string) bool) error {
	c, err := sarama.NewConsumer(addrs, nil)
	if err != nil {
		return err
	}

	pc, err := c.ConsumePartition(info.Topic, info.Partition, info.Offset)
	if err != nil {
		return err
	}

	defer func() {
		c.Close()
		pc.Close()
	}()

	l := info.End - info.Offset
	if l < end {
		end = l
	}

	for i := int64(0); i < end; i++ {
		select {
		case msg := <-pc.Messages():
			if stop := cb(string(msg.Value)); stop {
				return nil
			}
		case <-time.After(time.Second):
			break
		}
	}

	return nil
}
