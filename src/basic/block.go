package basic

import (
	"bytes"
	"crypto/sha256"
)

type Block struct {
	PrevHash []byte
	Hash     []byte
	Data     []byte
}

func (b *Block) CalculateHash() {
	// get hash of current block, hash = sha256(data + PrevHash)
	hash := sha256.Sum256(bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{' '}))
	b.Hash = hash[:]
}

func CreateBlock(_data string, _prevHash []byte) *Block {
	// create a new block with Hash value set
	newBlock := &Block{PrevHash: _prevHash, Data: []byte(_data)}
	newBlock.CalculateHash()
	return newBlock
}

func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}
