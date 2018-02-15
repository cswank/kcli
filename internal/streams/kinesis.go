package streams

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

type KinesisClient struct {
	cli *kinesis.Kinesis
}

func NewKinesis(region string) (*KinesisClient, error) {
	c := kinesis.New(session.New(&aws.Config{Region: aws.String(region)}))
	return &KinesisClient{cli: c}, nil
}

func (k *KinesisClient) GetTopics() ([]string, error) {
	resp, err := k.cli.ListStreams(nil)
	if err != nil {
		return nil, nil
	}

	out := make([]string, len(resp.StreamNames))
	for i, s := range resp.StreamNames {
		out[i] = *s
	}

	return out, nil
}

func (k *KinesisClient) GetTopic(streamName string) ([]Partition, error) {
	resp, err := k.cli.DescribeStream(&kinesis.DescribeStreamInput{StreamName: aws.String(streamName)})
	if err != nil {
		return nil, err
	}

	out := make([]Partition, len(resp.StreamDescription.Shards))
	for i, s := range resp.StreamDescription.Shards {
		out[i] = Partition{id: s.ShardId, stream: aws.String(streamName), Topic: streamName, Partition: int32(i)}
		//r := s.SequenceNumberRange
		log.Printf("%v\n", s)
	}

	return out, nil
}

func (k *KinesisClient) SearchTopic(partitions []Partition, term string, firstResult bool, cb func(int64, int64)) ([]Partition, error) {
	return nil, nil
}

func (k *KinesisClient) GetPartition(partition Partition, rows int, cb func(record []byte) bool) ([]Message, error) {
	i, err := k.cli.GetShardIterator(&kinesis.GetShardIteratorInput{
		ShardId:           partition.id,
		ShardIteratorType: aws.String("TRIM_HORIZON"),
		StreamName:        partition.stream,
	})
	if err != nil {
		return nil, err
	}

	// get records use shard iterator for making request
	records, err := k.cli.GetRecords(&kinesis.GetRecordsInput{
		ShardIterator: i.ShardIterator,
		Limit:         aws.Int64(int64(rows)),
	})

	if err != nil {
		return nil, err
	}

	out := make([]Message, len(records.Records))
	for i, r := range records.Records {
		out[i] = Message{
			Partition: partition,
			Value:     r.Data,
			Offset:    int64(i),
		}
	}

	return out, nil
}

func (k *KinesisClient) Search(partition Partition, term string, cb func(int64, int64)) (int64, error) {
	return 0, nil
}

func (k *KinesisClient) Fetch(partition Partition, end int64, cb func(string)) error {
	return nil
}

func (k *KinesisClient) Close() {}
