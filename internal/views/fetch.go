package views

import (
	"log"

	"github.com/cswank/kcli/internal/kafka"
)

//getTopics -> getTopic -> getPartition -> getMessage
func getTopics(size int, args string) (page, error) {
	r, err := kafka.GetTopics(size, args)
	log.Println("get topics", r, err)
	if err != nil {
		return page{}, err
	}

	return page{
		name:   "topics",
		header: "topics",
		body:   r,
		next:   getTopic,
	}, nil
}

func getTopic(size int, args string) (page, error) {
	r, err := kafka.GetTopic(size, args)
	if err != nil {
		return page{}, err
	}

	return page{
		name:   "topic",
		header: args,
		body:   r,
		next:   getPartition,
	}, nil
}

func getPartition(size int, args string) (page, error) {
	r, err := kafka.GetPartition(size, args)
	if err != nil {
		return page{}, err
	}

	return page{
		name:   "partition",
		header: args,
		body:   r,
		next:   getMessage,
	}, nil
}

func getMessage(size int, args string) (page, error) {
	r, err := kafka.GetMessage(size, args)
	if err != nil {
		return page{}, err
	}

	return page{
		name:   "message",
		header: args,
		body:   r,
	}, nil
}
