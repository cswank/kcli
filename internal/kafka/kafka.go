package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/cswank/kcli/internal/tunnel"
)

//Client fetches from kafka
type Client struct {
	addrs  []string
	sarama sarama.Client

	//when using ssh tunnels to connect
	hosts map[string]map[int32]string
	lock  sync.Mutex
}

//Partition holds information about a kafka partition
type Partition struct {
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
	Start     int64  `json:"start"`
	End       int64  `json:"end"`
	Offset    int64  `json:"offset"`
	Filter    string `json:"filter"`
}

//String turns a partition into a string
func (p *Partition) String() string {
	d, _ := json.Marshal(p)
	return string(d)
}

//Message holds information about a single kafka message
type Message struct {
	Partition Partition `json:"partition"`
	Value     []byte    `json:"msg"`
	Offset    int64     `json:"offset"`
}

//New returns a kafka Client.
func New(addrs []string, user string, port int) (*Client, error) {
	var hosts map[string]map[int32]string
	if user != "" {
		t := tunnel.New(user, port, addrs)
		var err error
		addrs, err = t.Connect()
		//TODO: figure out which node owns
		//which resource so a client can be created
		//for each host on the cluster.
		hosts = make(map[string]map[int32]string)
		if err != nil {
			return nil, err
		}
	}

	s, err := sarama.NewClient(addrs, nil)
	if err != nil {
		return nil, err
	}

	cli := &Client{
		sarama: s,
		addrs:  addrs,
		hosts:  hosts,
	}

	if user != "" {
		go cli.findLeaders(addrs)
	}

	return cli, nil
}

func (c *Client) findLeaders(addrs []string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	topics, err := c.sarama.Topics()
	if err != nil {
		if err != nil {
			log.Println("couldn't get topics")
			return
		}
	}

	for _, t := range topics {
		partitions, err := c.sarama.Partitions(t)
		if err != nil {
			log.Println("couldn't get partitions for topics", t, err)
			return
		}
		for _, p := range partitions {
			l, err := c.sarama.Leader(t, p)
			if err != nil {
				log.Printf("couldn't get leader for topic %s, partition %d: %s", t, p, err)
				return
			}

			addr := l.Addr()
			parts := strings.Split(addr, ":")
			var found string
			for _, a := range addrs {
				if strings.Index(a, parts[0]) == 0 {
					found = a
				}
			}

			if found == "" {
				log.Println("couldn't find a match address")
				return
			}
			m, ok := c.hosts[t]
			if !ok {
				m = map[int32]string{}
			}
			m[p] = found
			c.hosts[t] = m
		}
	}
}

//GetTopics gets topics (duh)
func (c *Client) GetTopics() ([]string, error) {
	return c.sarama.Topics()
}

//GetTopic gets a single kafka topic
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

//GetPartition fetches a kafka partition.  It includes a callback func
//so that the caller can tell it when to stop consuming.
func (c *Client) GetPartition(part Partition, end int, f func([]byte) bool) ([]Message, error) {
	var consumer sarama.Consumer
	var err error
	if c.hosts == nil {
		consumer, err = sarama.NewConsumer(c.addrs, nil)
	} else {
		c.lock.Lock()
		addr, ok := c.hosts[part.Topic][part.Partition]
		if !ok {
			c.lock.Unlock()
			return nil, fmt.Errorf("couldn't find address for topic %s and partition %d", part.Topic, part.Partition)
		}

		consumer, err = sarama.NewConsumer([]string{addr}, nil)
		c.lock.Unlock()
	}

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

	var out []Message

	var msg *sarama.ConsumerMessage
	var i int
	var last bool
	for i < end && !last {
		//log.Printf("i: %d, end: %d", i, end)
		select {
		case msg = <-pc.Messages():
			if f(msg.Value) {
				out = append(out, Message{
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

//Close disconnects from kafka
func (c *Client) Close() {
	c.sarama.Close()
}

type searchResult struct {
	partition Partition
	offset    int64
	error     error
}

//SearchTopic allows the caller to search across all partitions in a topic.
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

//Search is for searching for a string in a single kafka partition.
//It stops at the first match.
func (c *Client) Search(info Partition, s string, cb func(i, j int64)) (int64, error) {
	return c.search(info, s, func() bool { return false }, cb)
}

//Fetch gets all messages in a partition up intil the 'end' offset.
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
