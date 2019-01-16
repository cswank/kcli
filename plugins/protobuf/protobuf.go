package main

type Protobuf struct{}

func (p Protobuf) Decode(b []byte) []byte {
	return b
}

var Decoder Protobuf

func main() {}
