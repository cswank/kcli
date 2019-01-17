package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/Shopify/sarama"
	"github.com/cswank/kcli/dev/person"
	"github.com/golang/protobuf/proto"
)

var (
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
	rand.Seed(time.Now().UnixNano())
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
		x := item{
			First: randName(firstNames),
			Last:  randName(lastNames),
			Age:   randAge(),
		}

		y := &person.Person{
			First: x.First,
			Last:  x.Last,
			Age:   int32(x.Age),
		}

		l[i] = x
		d, _ := proto.Marshal(y)

		msg := &sarama.ProducerMessage{
			Topic: "proto",
			Key:   sarama.StringEncoder(x.First),
			Value: sarama.StringEncoder(string(d)),
		}

		_, _, err = producer.SendMessage(msg)
		if err != nil {
			log.Fatal(err)
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
