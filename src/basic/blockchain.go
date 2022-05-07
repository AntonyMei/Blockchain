package basic

type BlockChain struct {
	BlockList []*Block
}

func (bc *BlockChain) AddBlock(data string) {
	newBlock := CreateBlock(data, bc.BlockList[len(bc.BlockList)-1].Hash)
	bc.BlockList = append(bc.BlockList, newBlock)
}

func CreateBlockChain() *BlockChain {
	return &BlockChain{[]*Block{Genesis()}}
}
