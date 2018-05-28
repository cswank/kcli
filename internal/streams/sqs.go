package streams

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SQSClient struct {
	cli *sqs.SQS
}

func NewSQS(region string) (*SQSClient, error) {
	s := sqs.New(session.New(&aws.Config{Region: aws.String(region)}))
	return &SQSClient{cli: s}, nil
}

func (s *SQSClient) GetTopics() ([]string, error) {
	resp, err := s.cli.ListQueues(nil)
	if err != nil {
		return nil, nil
	}

	out := make([]string, len(resp.QueueUrls))
	for i, u := range resp.QueueUrls {
		out[i] = *u
	}

	return out, nil
}

func (k *SQSClient) GetTopic(streamName string) ([]Partition, error) {
	return nil, nil
}

func (s *SQSClient) SearchTopic(partitions []Partition, term string, firstResult bool, cb func(int64, int64)) ([]Partition, error) {
	return nil, nil
}

func (s *SQSClient) GetPartition(partition Partition, rows int, cb func(record []byte) bool) ([]Message, error) {
	return nil, nil
}

func (s *SQSClient) Search(partition Partition, term string, cb func(int64, int64)) (int64, error) {
	return 0, nil
}

func (s *SQSClient) Fetch(partition Partition, end int64, cb func(string)) error {
	return nil
}

func (s *SQSClient) Close() {}
