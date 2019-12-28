package main

import (
	"github.com/cswank/kcli/dev/person"
	"github.com/golang/protobuf/proto"
)

// Protobuf is an example of how to create a plugin
// to decode protobuf encoded kafka messages.
// To compile:
//     go build -buildmode=plugin -o protobuf.so main.go
// Then start kcli like:
//     kcli -d ./protobuf.so
type Protobuf struct{}

// Decode is required in order to be a plugin
func (p Protobuf) Decode(topic string, b []byte) ([]byte, error) {
	if topic != "proto" {
		return b, nil
	}

	var x person.Person
	err := proto.Unmarshal(b, &x)
	if err != nil {
		return nil, err
	}

	return []byte(x.String()), nil
}

// Decoder is the symbol that kcli will search for when
// using this as a plugin.
var Decoder Protobuf

func main() {}
