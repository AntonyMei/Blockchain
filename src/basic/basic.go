package basic

import (
	"bytes"
	"crypto/sha256"
)

type Block struct {
	prevHash []byte
	hash     []byte
	data     []byte
}

func (b *Block) CalculateHash() {
	// get hash of current block, hash = sha256(data + prevHash)
	hash := sha256.Sum256(bytes.Join([][]byte{b.data, b.prevHash}, []byte{' '}))
	b.hash = hash[:]
}

func CreateBlock(_data string, _prevHash []byte) *Block {
	// create a new block with hash value set
	newBlock := &Block{prevHash: _prevHash, data: []byte(_data)}
	newBlock.CalculateHash()
	return newBlock
}

type BlockChain struct {
	blockList []*Block
}
