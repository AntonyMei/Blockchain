package main

import (
	"fmt"
	"github.com/AntonyMei/Blockchain/src/blockchain"
	"github.com/AntonyMei/Blockchain/src/blocks"
	"strconv"
)

func main() {
	println("Persistent Storage Test\n")
	// create a chain
	chain := blockchain.InitBlockChain()
	chain.AddBlock("first block after genesis")
	chain.AddBlock("second block after genesis")
	chain.AddBlock("third block after genesis")

	// print info
	for iterator := chain.Iterator(); iterator.Next(); {
		block := iterator.GetVal()
		fmt.Printf("Previous hash: %x\n", block.PrevHash)
		fmt.Printf("hash: %x\n", block.Hash)
		fmt.Printf("data: %s\n", block.Data)
		fmt.Printf("nonce: %v\n", block.Nonce)
		fmt.Printf("nonce: %v\n", block.Difficulty)

		pow := blocks.CreateProofOfWork(block)
		fmt.Printf("Pow validated: %s\n", strconv.FormatBool(pow.ValidateNonce()))
		fmt.Println()
	}

}
