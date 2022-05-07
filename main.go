package main

import (
	"fmt"
	"github.com/AntonyMei/Blockchain/src/basic"
	"strconv"
)

func main() {
	println("Blockchain Chapter 2")
	// create a chain
	chain := basic.CreateBlockChain()
	chain.AddBlock("first block after genesis")
	chain.AddBlock("second block after genesis")
	chain.AddBlock("third block after genesis")

	// print info
	for _, block := range chain.BlockList {
		fmt.Printf("Previous hash: %x\n", block.PrevHash)
		fmt.Printf("hash: %x\n", block.Hash)
		fmt.Printf("data: %s\n", block.Data)
		fmt.Printf("nonce: %v\n", block.Nonce)
		fmt.Printf("nonce: %v\n", block.Difficulty)

		pow := basic.CreateProofOfWork(block)
		fmt.Printf("Pow validated: %s\n", strconv.FormatBool(pow.ValidateNonce()))
		fmt.Println()
	}

}
