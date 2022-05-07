package basic

import pow2 "github.com/AntonyMei/Blockchain/src/pow"

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
	pow := pow2.CreateProofOfWork(newBlock)
	nonce, hash := pow.GenerateNonce()
	newBlock.Nonce = nonce
	newBlock.Hash = hash
	return newBlock
}

func Genesis(_difficulty int) *Block {
	return CreateBlock("Genesis", []byte{}, _difficulty)
}
