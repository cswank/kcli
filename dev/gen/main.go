package main

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/Shopify/sarama"
)

var (
	firstNames = []string{"Hector", "Orlando", "Marc", "Tami", "Jorge", "Leigh", "Sabrina", "Jim", "Elaine", "Jeanne", "Felix", "Marco", "Kelvin", "Owen", "Debra", "Elizabeth", "Kevin", "Jane", "Marlene", "Fannie", "Frank", "Frances", "Cheryl", "Yolanda", "Troy", "Shannon", "Donnie", "Ginger", "Ashley", "Danny", "Kristi", "Angela", "Sylvester", "Robert", "Richard", "Maggie", "Eduardo", "Randolph", "Monica", "Lucia", "Sandy", "Elias", "Bonnie", "Heather", "Abraham", "Darnell", "Rudolph", "Leah", "Daisy", "Doris", "Kendra", "Bryant", "Bobbie", "Shelley", "Bob", "Rosalie", "Kerry", "Susan", "Natasha", "Winifred", "Marty", "Ernestine", "Sadie", "Dave", "Katherine", "Don", "Mattie", "Mario", "Donald", "Arturo", "Arlene", "Cecilia", "Joyce", "Kelley", "Carol", "Ed", "Roy", "Sara", "Pauline", "Blanca", "Dianna", "Candice", "Mark", "Patricia", "Steve", "Roxanne", "Clifford", "Sheri", "Carroll", "Latoya", "Martha", "Kent", "Jacqueline", "Grace", "Jeremiah", "Travis", "Ron", "Angelica", "Verna", "Tabitha"}
	lastNames  = []string{"Gray", "Robertson", "Johnston", "Gregory", "Sullivan", "Powers", "Nash", "Rowe", "Parks", "Fleming", "Paul", "Pratt", "Scott", "Maldonado", "Moody", "Myers", "Brock", "Becker", "Wheeler", "Greene", "Anderson", "Mack", "Hawkins", "Wilkerson", "Burton", "Payne", "Quinn", "Drake", "Rivera", "Norman", "Banks", "Rice", "Flores", "Schneider", "Morris", "Nelson", "Wallace", "Green", "Ramirez", "Medina", "Price", "West", "Kim", "Schultz", "Pearson", "King", "Lopez", "Watts", "Patterson", "Garza", "Daniel", "Carter", "Andrews", "Lloyd", "Garner", "Carpenter", "Simpson", "Jones", "Buchanan", "Lynch", "Holmes", "Horton", "Walton", "Wilson", "Stone", "Cortez", "Harrison", "Bates", "Lucas", "Cohen", "Zimmerman", "Gross", "Cain", "Doyle", "Fitzgerald", "Alvarado", "Cooper", "Sims", "Dean", "Hines", "Jensen", "Johnson", "Hicks", "Stephens", "Singleton", "Hardy", "Shaw", "Russell", "Beck", "Austin", "Olson", "Obrien", "Porter", "Lee", "Howard", "Steele", "Bass", "Mills", "Crawford", "Padilla"}

	topics = map[string]string{
		"stuff":   `{"first_name": "%s", "last_name": "%s", "age": %d}`,
		"things":  `first_name: %s, last_name: %s, age: %d`,
		"whatnot": `{"first_name": "%s", "last_name": "%s", "age": %d}`,
	}
)

func main() {
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
		for i := 0; i < 500; i++ {
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
}

func randName(names []string) string {
	return names[rand.Intn(len(names))]
}

func randAge() int {
	return 5 + rand.Intn(80)
}
