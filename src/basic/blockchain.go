package basic

// InitialChainDifficulty / 4 = # of zeros at hash head
const InitialChainDifficulty = 16

type BlockChain struct {
	BlockList       []*Block
	ChainDifficulty int
}

func CreateBlockChain() *BlockChain {
	return &BlockChain{BlockList: []*Block{Genesis(InitialChainDifficulty)},
		ChainDifficulty: InitialChainDifficulty}
}

func (bc *BlockChain) AddBlock(data string) {
	newBlock := CreateBlock(data, bc.BlockList[len(bc.BlockList)-1].Hash, bc.ChainDifficulty)
	bc.BlockList = append(bc.BlockList, newBlock)
}
