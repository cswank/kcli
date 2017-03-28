package kafka

import (
	"fmt"
)

func getMockTopics(size int, args string) ([][]Row, error) {
	r := make([]Row, size)
	for i := 0; i < size; i++ {
		s := fmt.Sprintf("topic %d", i)
		r[i] = Row{Args: s, Data: s}
	}
	return [][]Row{r}, nil
}

func getMockTopic(size int, args string) ([][]Row, error) {
	return [][]Row{
		[]Row{
			{Args: "partition 1", Data: "partition 1"},
			{Args: "partition 2", Data: "partition 2"},
			{Args: "partition 3", Data: "partition 3"},
		},
	}, nil
}

func getMockPartition(size int, args string) ([][]Row, error) {
	return [][]Row{
		[]Row{
			{Args: `{"offset": 0}`, Data: `{"name": "fred"}`},
			{Args: `{"offset": 1}`, Data: `{"name": "craig"}`},
			{Args: `{"offset": 2}`, Data: `{"name": "laura"}`},
		},
	}, nil
}

func getMockMessage(size int, args string) ([][]Row, error) {
	return [][]Row{
		[]Row{
			{Data: `{"name": "fred"}`},
		},
	}, nil
}
