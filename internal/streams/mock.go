package streams

import (
	"fmt"
	"math/rand"
)

var (
	names = []string{"Burl", "Bernetta", "Kelsey", "Lieselotte", "Mechelle", "Migdalia", "Mammie", "Hiroko", "Dalia", "Janelle", "Elyse", "Barb", "Major", "Stacey", "Edda", "Theola", "Queenie", "Deedee", "Marya", "Addie", "Joye", "Klara", "Robbie", "Timika", "Wendy", "Gemma", "Helen", "Yen", "Gena", "Kathlene", "Jule", "Lani", "Enriqueta", "Laci", "Georgie", "Nana", "Kori", "Maryam", "Dominica", "Cheree", "Garnett", "Gearldine", "Branda", "Amada", "Darlena", "Keena", "Rosemary", "Stacia", "Gayla", "So"}
)

func getMockTopics() ([]string, error) {
	var t []string
	for i := 0; i < 10; i++ {
		t = append(t, fmt.Sprintf("topic %d", i))
	}
	return t, nil
}

func getMockTopic(topic string) ([]Partition, error) {
	p := make([]Partition, 100)
	for i := 0; i < 100; i++ {
		s := rand.Intn(5000)
		e := rand.Intn(5000)
		p[i] = Partition{Partition: int32(i), Topic: topic, Start: int64(s), End: int64(s + e)}
	}
	return p, nil
}

func getMockPartition(part Partition, num int) ([]Message, error) {
	end := 100 + rand.Intn(5000)
	m := make([]Message, end)
	for i := 0; i < end; i++ {
		m[i] = Message{
			Value: []byte(fmt.Sprintf(`{"name": "%s", "age": %d}`, names[rand.Intn(len(names))], 1+rand.Intn(100))),
			Partition: Partition{
				Partition: part.Partition,
				Topic:     part.Topic,
				Start:     0,
				End:       int64(end),
				Offset:    int64(i),
			},
		}
	}
	return m, nil
}

func mockFetch(_ Partition, _ int64, cb func(string)) error {
	for i := 0; i < 10; i++ {
		cb(fmt.Sprintf("%d", i))
	}

	return nil
}

func mockSearch(info Partition, s string) (int64, error) {
	return int64(-1), nil
}
