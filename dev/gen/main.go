package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/Shopify/sarama"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	kingpin "gopkg.in/alecthomas/kingpin.v1"
)

var (
	region = kingpin.Arg("region", "aws region for kinesis").String()

	firstNames = []string{"Hector", "Orlando", "Marc", "Tami", "Jorge", "Leigh", "Sabrina", "Jim", "Elaine", "Jeanne", "Felix", "Marco", "Kelvin", "Owen", "Debra", "Elizabeth", "Kevin", "Jane", "Marlene", "Fannie", "Frank", "Frances", "Cheryl", "Yolanda", "Troy", "Shannon", "Donnie", "Ginger", "Ashley", "Danny", "Kristi", "Angela", "Sylvester", "Robert", "Richard", "Maggie", "Eduardo", "Randolph", "Monica", "Lucia", "Sandy", "Elias", "Bonnie", "Heather", "Abraham", "Darnell", "Rudolph", "Leah", "Daisy", "Doris", "Kendra", "Bryant", "Bobbie", "Shelley", "Bob", "Rosalie", "Kerry", "Susan", "Natasha", "Winifred", "Marty", "Ernestine", "Sadie", "Dave", "Katherine", "Don", "Mattie", "Mario", "Donald", "Arturo", "Arlene", "Cecilia", "Joyce", "Kelley", "Carol", "Ed", "Roy", "Sara", "Pauline", "Blanca", "Dianna", "Candice", "Mark", "Patricia", "Steve", "Roxanne", "Clifford", "Sheri", "Carroll", "Latoya", "Martha", "Kent", "Jacqueline", "Grace", "Jeremiah", "Travis", "Ron", "Angelica", "Verna", "Tabitha"}
	lastNames  = []string{"Gray", "Robertson", "Johnston", "Gregory", "Sullivan", "Powers", "Nash", "Rowe", "Parks", "Fleming", "Paul", "Pratt", "Scott", "Maldonado", "Moody", "Myers", "Brock", "Becker", "Wheeler", "Greene", "Anderson", "Mack", "Hawkins", "Wilkerson", "Burton", "Payne", "Quinn", "Drake", "Rivera", "Norman", "Banks", "Rice", "Flores", "Schneider", "Morris", "Nelson", "Wallace", "Green", "Ramirez", "Medina", "Price", "West", "Kim", "Schultz", "Pearson", "King", "Lopez", "Watts", "Patterson", "Garza", "Daniel", "Carter", "Andrews", "Lloyd", "Garner", "Carpenter", "Simpson", "Jones", "Buchanan", "Lynch", "Holmes", "Horton", "Walton", "Wilson", "Stone", "Cortez", "Harrison", "Bates", "Lucas", "Cohen", "Zimmerman", "Gross", "Cain", "Doyle", "Fitzgerald", "Alvarado", "Cooper", "Sims", "Dean", "Hines", "Jensen", "Johnson", "Hicks", "Stephens", "Singleton", "Hardy", "Shaw", "Russell", "Beck", "Austin", "Olson", "Obrien", "Porter", "Lee", "Howard", "Steele", "Bass", "Mills", "Crawford", "Padilla"}

	topics = map[string]string{
		"stuff":   `{"first_name": "%s", "last_name": "%s", "age": %d}`,
		"things":  `first_name: %s, last_name: %s, age: %d`,
		"whatnot": `{"first_name": "%s", "last_name": "%s", "age": %d}`,
		"items":   `%s,%s,%d`,
		"other":   `{"first_name": "%s", "last_name": "%s", "age": %d}`,
	}
)

func main() {
	kingpin.Parse()
	rand.Seed(time.Now().UnixNano())
	if *region != "" {
		writeKinesis(*region)
	} else {
		writeKafka()
	}
}

func writeKinesis(region string) {
	s := session.New(&aws.Config{Region: aws.String(region)})
	kc := kinesis.New(s)

	i := int64(2)
	for topic, tpl := range topics {
		out, err := kc.CreateStream(&kinesis.CreateStreamInput{
			ShardCount: aws.Int64(i),
			StreamName: aws.String(topic),
		})

		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v\n", out)

		if err := kc.WaitUntilStreamExists(&kinesis.DescribeStreamInput{StreamName: aws.String(topic)}); err != nil {
			log.Fatal(err)
		}

		for i := 0; i < 500; i++ {
			first := randName(firstNames)
			last := randName(lastNames)
			age := randAge()

			val := fmt.Sprintf(tpl, first, last, age)

			_, err := kc.PutRecord(&kinesis.PutRecordInput{
				Data:         []byte(val),
				StreamName:   aws.String(topic),
				PartitionKey: aws.String(first),
			})

			if err != nil {
				log.Fatal(err)
			}
		}
		i++
	}
}

func writeKafka() {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	brokers := []string{"localhost:9092"}
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	for topic, tpl := range topics {
		for i := 0; i < 5000; i++ {
			first := randName(firstNames)
			last := randName(lastNames)
			age := randAge()

			val := fmt.Sprintf(tpl, first, last, age)

			msg := &sarama.ProducerMessage{
				Topic: topic,
				Key:   sarama.StringEncoder(last),
				Value: sarama.StringEncoder(val),
			}

			_, _, err := producer.SendMessage(msg)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	l := make([]item, 400)
	for i := 0; i < 400; i++ {
		l[i] = item{
			First: randName(firstNames),
			Last:  randName(lastNames),
			Age:   randAge(),
		}
	}

	d, _ := json.Marshal(l)

	msg := &sarama.ProducerMessage{
		Topic: "miscellany",
		Key:   sarama.StringEncoder("item"),
		Value: sarama.StringEncoder(string(d)),
	}
	_, _, err = producer.SendMessage(msg)
	if err != nil {
		log.Fatal(err)
	}
}

type item struct {
	First string `json:"first"`
	Last  string `json:"last"`
	Age   int    `json:"name"`
}

func randName(names []string) string {
	return names[rand.Intn(len(names))]
}

func randAge() int {
	return 5 + rand.Intn(80)
}
