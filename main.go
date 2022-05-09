package main

import (
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
	tx1 := chain.GenerateTransaction("Alice", "Bob", 30)
	chain.AddBlock("Bob", "Third Block after genesis", []*transaction.Transaction{tx1})
	// Alice gives Bob 10, then 40, then Bob returns 60, Charlie logs this
	tx2 := chain.GenerateTransaction("Alice", "Bob", 10)
	tx3 := chain.GenerateTransaction("Alice", "Bob", 40)
	tx4 := chain.GenerateTransaction("Bob", "Alice", 60)
	chain.AddBlock("Charlie", "Fourth Block after genesis",
		[]*transaction.Transaction{tx2, tx3, tx4})

	// print info
	chain.Log2Terminal()

}
