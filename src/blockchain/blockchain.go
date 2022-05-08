package blockchain

import (
	"github.com/AntonyMei/Blockchain/src/blocks"
)

// InitialChainDifficulty / 4 = # of zeros at hash head
const InitialChainDifficulty = 16

type BlockChain struct {
	BlockList       []*blocks.Block
	ChainDifficulty int
}

func CreateBlockChain() *BlockChain {
	return &BlockChain{BlockList: []*blocks.Block{blocks.Genesis(InitialChainDifficulty)},
		ChainDifficulty: InitialChainDifficulty}
}

func (bc *BlockChain) AddBlock(data string) {
	newBlock := blocks.CreateBlock(data, bc.BlockList[len(bc.BlockList)-1].Hash, bc.ChainDifficulty)
	bc.BlockList = append(bc.BlockList, newBlock)
}
