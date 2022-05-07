package basic

const InitialChainDifficulty = 12

type BlockChain struct {
	BlockList       []*Block
	ChainDifficulty int
}

func (bc *BlockChain) AddBlock(data string) {
	newBlock := CreateBlock(data, bc.BlockList[len(bc.BlockList)-1].Hash, bc.ChainDifficulty)
	bc.BlockList = append(bc.BlockList, newBlock)
}

func CreateBlockChain() *BlockChain {
	return &BlockChain{BlockList: []*Block{Genesis(InitialChainDifficulty)},
		ChainDifficulty: InitialChainDifficulty}
}
