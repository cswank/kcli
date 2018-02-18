package streams

type Streamer interface {
	Source() string
	GetTopics() ([]string, error)
	GetTopic(topic string) ([]Partition, error)
	SearchTopic(partitions []Partition, term string, firstResult bool, cb func(int64, int64)) ([]Partition, error)
	GetPartition(partition Partition, rows int, cb func(record []byte) bool) ([]Message, error)
	Search(partition Partition, term string, cb func(int64, int64)) (int64, error)
	Fetch(partition Partition, end int64, cb func(string)) error
	Close()
}
