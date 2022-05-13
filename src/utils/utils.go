package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/mr-tron/base58"
	"log"
	"os"
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
	DoubleSpending
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
	case DoubleSpending:
		return "DoubleSpending"
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

func Match(inputList []string, expected []string) bool {
	if len(inputList) < len(expected) {
		return false
	}
	for idx := range expected {
		if inputList[idx] != expected[idx] {
			return false
		}
	}
	return true
}

func CheckArgumentCount(inputList []string, expected int) bool {
	if len(inputList) != expected {
		fmt.Printf("Expect %v arguments, got %v instead.\n", expected, len(inputList))
		return false
	}
	return true
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
