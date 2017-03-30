package kafka

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/Shopify/sarama"
)

var (
	GetTopics    func() ([]string, error)
	GetTopic     func(string) ([]Partition, error)
	GetPartition func(Partition, int) ([]Msg, error)
	Fetch        func(Partition, int64, func(string)) error
	Search       func(Partition, string) (int64, error)

	addrs []string
	cli   sarama.Client
)

type Partition struct {
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
	Start     int64  `json:"start"`
	End       int64  `json:"end"`
	Offset    int64  `json:"offset"`
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

func init() {
	GetTopics = getMockTopics
	GetTopic = getMockTopic
	GetPartition = getMockPartition
	Fetch = mockFetch
	Search = mockSearch
}

func Connect(a []string) {
	addrs = a
	var err error
	cli, err = sarama.NewClient(addrs, nil)
	if err != nil {
		log.Fatal(err)
	}

	GetTopics = cli.Topics
	GetTopic = getTopic
	GetPartition = getPartition
	Fetch = fetch
	Search = search
}

func getTopic(topic string) ([]Partition, error) {
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

func getPartition(part Partition, end int) ([]Msg, error) {
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

	l := int(part.End - part.Offset)
	if l < end {
		end = l
	}

	var out []Msg

	var msg *sarama.ConsumerMessage
	for i := 0; i < end; i++ {
		select {
		case msg = <-pc.Messages():
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
		case <-time.After(time.Second):
			break
		}
	}

	log.Println("consuming kafka", part.Topic, part.Partition, part.Offset, part.End, end, len(out))
	return out, nil
}

func Close() {
	if cli != nil {
		cli.Close()
	}
}

func search(info Partition, s string) (int64, error) {
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

func fetch(info Partition, end int64, cb func(string)) error {
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
				break
			}
		case <-time.After(time.Second):
			break
		}
	}

	return nil
}
