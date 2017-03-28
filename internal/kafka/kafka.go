package kafka

type Row struct {
	Args string
	Data string
}

type fetcher func(int, string) [][]Row

var (
	GetTopics    = getMockTopics
	GetTopic     = getMockTopic
	GetPartition = getMockPartition
	GetMessage   = getMockMessage
)
