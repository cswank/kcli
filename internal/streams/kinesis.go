package streams

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

type KinesisClient struct {
	cli         *kinesis.Kinesis
	first, last string
}

func NewKinesis(region string) (*KinesisClient, error) {
	c := kinesis.New(session.New(&aws.Config{Region: aws.String(region)}))
	return &KinesisClient{cli: c}, nil
}

func (k *KinesisClient) Source() string { return "kinesis" }

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
		out[i] = Partition{id: s.ShardId, stream: aws.String(streamName), Topic: streamName, Partition: int32(i), Start: 0, End: 2}
		//r := s.SequenceNumberRange
	}

	return out, nil
}

func (k *KinesisClient) SearchTopic(partitions []Partition, term string, firstResult bool, cb func(int64, int64)) ([]Partition, error) {
	return nil, nil
}

func (k *KinesisClient) GetPartition(partition Partition, rows int, cb func(record []byte) bool) ([]Message, error) {
	var ii *kinesis.GetShardIteratorInput
	if partition.After != nil {
		ii = &kinesis.GetShardIteratorInput{
			ShardId:           partition.id,
			ShardIteratorType: aws.String("AT_TIMESTAMP"),
			StreamName:        partition.stream,
			Timestamp:         partition.After,
		}
	} else if partition.SequenceNumber != nil {
		ii = &kinesis.GetShardIteratorInput{
			ShardId:                partition.id,
			ShardIteratorType:      aws.String("AFTER_SEQUENCE_NUMBER"),
			StreamName:             partition.stream,
			StartingSequenceNumber: partition.SequenceNumber,
		}
	} else {
		ii = &kinesis.GetShardIteratorInput{
			ShardId:           partition.id,
			ShardIteratorType: aws.String("TRIM_HORIZON"),
			StreamName:        partition.stream,
		}
	}
	x, err := k.cli.GetShardIterator(ii)
	if err != nil {
		return nil, err
	}

	iter := x.ShardIterator
	// get records use shard iterator for making request
	var out []Message
	var i int
	for {
		limit := int64(rows - i)
		if limit <= 0 {
			break
		}

		records, err := k.cli.GetRecords(&kinesis.GetRecordsInput{
			ShardIterator: iter,
			Limit:         aws.Int64(limit),
		})

		log.Printf("got %d records, err: %v", len(records.Records), err)
		if err != nil {
			return nil, err
		}
		if len(records.Records) == 0 {
			break
		}

		for _, r := range records.Records {
			out = append(out, Message{
				Partition:      partition,
				Value:          r.Data,
				Offset:         int64(i),
				Timestamp:      r.ApproximateArrivalTimestamp,
				SequenceNumber: r.SequenceNumber,
			})
			i++
		}
		iter = records.NextShardIterator
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
