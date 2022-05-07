package main

import (
	"fmt"
	"github.com/AntonyMei/Blockchain/src/basic"
)

func main() {
	println("Blockchain Chapter 1")
	chain := basic.CreateBlockChain()
	chain.AddBlock("first block")
	chain.AddBlock("second block")
	for _, block := range chain.BlockList {
		fmt.Printf("Previous hash: %x\n", block.PrevHash)
		fmt.Printf("data: %s\n", block.Data)
		fmt.Printf("hash: %x\n", block.Hash)
	}
}
