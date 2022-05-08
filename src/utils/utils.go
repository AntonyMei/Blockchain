package utils

import (
	"bytes"
	"encoding/binary"
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
