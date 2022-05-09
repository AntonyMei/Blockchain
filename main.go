package main

import (
	"fmt"
	"github.com/AntonyMei/Blockchain/src/blockchain"
	"github.com/AntonyMei/Blockchain/src/transaction"
)

func main() {
	println("Transaction Test")
	// alice starts a chain / continues from last chain
	chain := blockchain.InitBlockChain("Alice")
	// then mines a block
	chain.AddBlock("Alice", "First Block after genesis", []*transaction.Transaction{})
	// bob comes in and mine another block
	chain.AddBlock("Bob", "Second Block after genesis", []*transaction.Transaction{})
	// Alice pay bob 30 in the next block
	tx1 := chain.GenerateTransaction("Alice", []string{"Bob"}, []int{30})
	chain.AddBlock("Bob", "Third Block after genesis", []*transaction.Transaction{tx1})
	// Alice gives Bob 90, David 40, then Bob returns 60, Charlie logs this
	tx2 := chain.GenerateTransaction("Alice", []string{"Bob", "David"}, []int{90, 40})
	tx3 := chain.GenerateTransaction("Bob", []string{"Alice"}, []int{60})
	chain.AddBlock("Charlie", "Fourth Block after genesis",
		[]*transaction.Transaction{tx2, tx3})
	// At this point the balance should look like
	// Alice:   100
	// Bob:     260
	// Charlie: 100
	// David:   40
	// Total:   500

	// print info
	chain.Log2Terminal()
	fmt.Printf("Final Balance\n")
	fmt.Printf("Alice: %v.\n", chain.GetBalance("Alice"))
	fmt.Printf("Bob: %v.\n", chain.GetBalance("Bob"))
	fmt.Printf("Charlie: %v.\n", chain.GetBalance("Charlie"))
	fmt.Printf("David: %v.\n", chain.GetBalance("David"))
	fmt.Printf("Eta: %v.\n", chain.GetBalance("Eta"))
}
