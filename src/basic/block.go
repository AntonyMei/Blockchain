package basic

import (
	"bytes"
	"encoding/gob"
	"github.com/AntonyMei/Blockchain/src/utils"
)

type Block struct {
	// basic
	PrevHash []byte
	Hash     []byte
	Data     []byte
	// proof of work
	Nonce      int
	Difficulty int
}

func CreateBlock(_data string, _prevHash []byte, _difficulty int) *Block {
	// create block with given data and difficulty
	newBlock := &Block{PrevHash: _prevHash, Hash: []byte{}, Data: []byte(_data),
		Nonce: 0, Difficulty: _difficulty}
	pow := CreateProofOfWork(newBlock)
	nonce, hash := pow.GenerateNonceHash()
	newBlock.Nonce = nonce
	newBlock.Hash = hash
	return newBlock
}

func Genesis(_difficulty int) *Block {
	return CreateBlock("Genesis", []byte{}, _difficulty)
}

func (b *Block) Serialize() []byte {
	// serialize a block into byte stream
	var result bytes.Buffer
	var encoder = gob.NewEncoder(&result)
	utils.Handle(encoder.Encode(b))
	return result.Bytes()
}

func Deserialize(stream []byte) *Block {
	// deserialize byte stream from block
	var block Block
	var decoder = gob.NewDecoder(bytes.NewReader(stream))
	utils.Handle(decoder.Decode(&block))
	return &block
}
