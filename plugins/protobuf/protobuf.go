package main

import (
	"github.com/cswank/kcli/dev/person"
	"github.com/golang/protobuf/proto"
)

// Example of how to create a protobuf plugin for your data structure
type Protobuf struct{}

func (p Protobuf) Decode(b []byte) ([]byte, error) {
	var x person.Person
	err := proto.Unmarshal(b, &x)
	if err != nil {
		return nil, err
	}

	return []byte(x.String()), nil
}

var Decoder Protobuf

func main() {}
