package utils

import (
	"bytes"
	"encoding/binary"
	"github.com/mr-tron/base58"
	"log"
)

type BlockStatus int64

const (
	Verified = iota
	WrongGenesis
	PrevBlockNotFound
	HashMismatch
	WrongTxID
	TooManyCoinbaseTX
	SourceTXONotFound
	WrongTXInputSignature
	InputSumOutputSumMismatch
)

func (bs BlockStatus) String() string {
	switch bs {
	case Verified:
		return "Verified"
	case WrongGenesis:
		return "WrongGenesis"
	case PrevBlockNotFound:
		return "PrevBlockNotFound"
	case HashMismatch:
		return "HashMismatch"
	case WrongTxID:
		return "WrongTxID"
	case TooManyCoinbaseTX:
		return "TooManyCoinbaseTX"
	case SourceTXONotFound:
		return "SourceTXONotFound"
	case WrongTXInputSignature:
		return "WrongTXInputSignature"
	case InputSumOutputSumMismatch:
		return "InputSumOutputSumMismatch"
	}
	return "Unknown"
}

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
