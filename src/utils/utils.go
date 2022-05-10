package utils

import (
	"bytes"
	"encoding/binary"
	"github.com/mr-tron/base58"
	"log"
)

func Int2Hex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

func Base58Encode(input []byte) []byte {
	encode := base58.Encode(input)
	return []byte(encode)
}

func Base58Decode(input []byte) []byte {
	decode, err := base58.Decode(string(input[:]))
	Handle(err)
	return decode
}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func Assert(exp bool, msg string) {
	if !exp {
		log.Panic("Assertion failed: " + msg)
	}
}
