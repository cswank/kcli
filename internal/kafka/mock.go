package kafka

import (
	"fmt"
)

func getMockTopics() ([]string, error) {
	var t []string
	for i := 0; i < 10; i++ {
		t = append(t, fmt.Sprintf("topic %d", i))
	}
	return t, nil
}

func getMockTopic(topic string) ([]Partition, error) {
	return []Partition{
		{Partition: 1, Topic: topic, Start: 1243, End: 12001},
		{Partition: 2, Topic: topic, Start: 0, End: 99},
		{Partition: 3, Topic: topic, Start: 4004, End: 5005},
	}, nil
}

func getMockPartition(part Partition, num int) ([]Msg, error) {
	return []Msg{
		{Value: []byte(`{"name": "craig"}`), Partition: Partition{Partition: part.Partition, Topic: part.Topic, Start: 1243, End: 12001}},
		{Value: []byte(`{"name": "james"}`), Partition: Partition{Partition: part.Partition, Topic: part.Topic, Start: 0, End: 99}},
		{Value: []byte(`{"name": "ronnie"}`), Partition: Partition{Partition: part.Partition, Topic: part.Topic, Start: 4004, End: 5005}},
		{Value: []byte(`{"name": "jeff"}`), Partition: Partition{Partition: part.Partition, Topic: part.Topic, Start: 4, End: 11}},
	}, nil
}

func mockFetch(_ Partition, _ int64, cb func(string)) error {
	for i := 0; i < 10; i++ {
		cb(fmt.Sprintf("%d", i))
	}

	return nil
}

func mockSearch(info Partition, s string) (int64, error) {
	return int64(-1), nil
}
