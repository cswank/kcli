package kafka

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Shopify/sarama"
)

type Client struct {
	addrs  []string
	sarama sarama.Client
}

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

func New(addrs []string, tunnel bool) (*Client, error) {
	s, err := sarama.NewClient(addrs, nil)
	if err != nil {
		return nil, err
	}

	cli := &Client{
		sarama: s,
		addrs:  addrs,
	}

	if tunnel {
		brokers := cli.sarama.Brokers()
		for _, b := range brokers {
			fmt.Printf("broker: %+v\n", b)
		}
		// topics, err := cli.Topics()
		// if err != nil {
		// 	return err
		// }
	}
	return cli, nil
}

func (c *Client) GetTopics() ([]string, error) {
	return c.sarama.Topics()
}

func (c *Client) GetTopic(topic string) ([]Partition, error) {
	partitions, err := c.sarama.Partitions(topic)
	if err != nil {
		return nil, err
	}

	out := make([]Partition, len(partitions))

	for i, p := range partitions {
		n, err := c.sarama.GetOffset(topic, p, sarama.OffsetNewest)
		if err != nil {
			return nil, err
		}

		o, err := c.sarama.GetOffset(topic, p, sarama.OffsetOldest)
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

func (c *Client) GetPartition(part Partition, end int, f func([]byte) bool) ([]Msg, error) {
	consumer, err := sarama.NewConsumer(c.addrs, nil)
	if err != nil {
		return nil, err
	}

	pc, err := consumer.ConsumePartition(part.Topic, part.Partition, part.Offset)
	if err != nil {
		return nil, err
	}

	defer func() {
		consumer.Close()
		pc.Close()
	}()

	var out []Msg

	var msg *sarama.ConsumerMessage
	var i int
	var last bool
	for i < end && !last {
		//log.Printf("i: %d, end: %d", i, end)
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

func (c *Client) Close() {
	c.sarama.Close()
}

type searchResult struct {
	partition Partition
	offset    int64
	error     error
}

func (c *Client) SearchTopic(partitions []Partition, s string, firstResult bool, cb func(int64, int64)) ([]Partition, error) {
	ch := make(chan searchResult)
	n := int64(len(partitions))
	var stop bool
	f := func() bool {
		return stop
	}

	for _, p := range partitions {
		go func(partition Partition, ch chan searchResult) {
			i, err := c.search(partition, s, f, func(_, _ int64) {})
			ch <- searchResult{partition: partition, offset: i, error: err}
		}(p, ch)
	}

	var results []Partition

	nResults := len(partitions)
	if firstResult {
		nResults = 1
	}

	var i int64
	for i = 0; i < int64(len(partitions)); i++ {
		r := <-ch
		cb(i, n)
		if r.error != nil {
			return nil, r.error
		}
		if r.offset > -1 {
			r.partition.Offset = r.offset
			results = append(results, r.partition)
		}
		if len(results) == nResults {
			stop = true
			break
		}

	}

	sort.Slice(results, func(i, j int) bool {
		return results[j].Partition >= results[i].Partition
	})

	return results, nil
}

func (c *Client) search(info Partition, s string, stop func() bool, cb func(int64, int64)) (int64, error) {
	n := int64(-1)
	var i int64
	err := c.consume(info, info.End, func(msg string) bool {
		cb(i, info.End)
		if strings.Contains(msg, s) {
			n = i + info.Offset
			return true
		}
		i++
		return stop()
	})

	return n, err
}

func (c *Client) Search(info Partition, s string, cb func(i, j int64)) (int64, error) {
	return c.search(info, s, func() bool { return false }, cb)
}

func (c *Client) Fetch(info Partition, end int64, cb func(string)) error {
	return c.consume(info, end, func(s string) bool {
		cb(s)
		return false
	})
}

func (c *Client) consume(info Partition, end int64, cb func(string) bool) error {
	consumer, err := sarama.NewConsumer(c.addrs, nil)
	if err != nil {
		return err
	}

	pc, err := consumer.ConsumePartition(info.Topic, info.Partition, info.Offset)
	if err != nil {
		return err
	}

	defer func() {
		consumer.Close()
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
