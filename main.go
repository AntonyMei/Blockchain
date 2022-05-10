package main

import (
	"fmt"
	"github.com/AntonyMei/Blockchain/src/blockchain"
	"github.com/AntonyMei/Blockchain/src/transaction"
	"github.com/AntonyMei/Blockchain/src/wallet"
)

func main() {
	println("Wallet Test")
	// initialize wallets
	wallets, err := wallet.InitializeWallets()
	var aliceAddr, bobAddr, charlieAddr, davidAddr []byte
	var aliceWallet, bobWallet, charlieWallet, davidWallet *wallet.Wallet
	if err != nil {
		aliceAddr = wallets.CreateWallet("Alice")
		aliceWallet = wallets.GetWallet("Alice")
		bobAddr = wallets.CreateWallet("Bob")
		bobWallet = wallets.GetWallet("Bob")
		charlieAddr = wallets.CreateWallet("Charlie")
		charlieWallet = wallets.GetWallet("Charlie")
		davidAddr = wallets.CreateWallet("David")
		davidWallet = wallets.GetWallet("David")
	} else {
		aliceWallet = wallets.GetWallet("Alice")
		aliceAddr = aliceWallet.Address()
		bobWallet = wallets.GetWallet("Bob")
		bobAddr = bobWallet.Address()
		charlieWallet = wallets.GetWallet("Charlie")
		charlieAddr = charlieWallet.Address()
		davidWallet = wallets.GetWallet("David")
		davidAddr = davidWallet.Address()
	}

	// alice starts a chain / continues from last chain
	chain := blockchain.InitBlockChain(aliceAddr)
	// then mines a block
	chain.AddBlock(aliceAddr, "First Block after genesis", []*transaction.Transaction{})
	// bob comes in and mine another block
	chain.AddBlock(bobAddr, "Second Block after genesis", []*transaction.Transaction{})
	// Alice pay bob 30 in the next block
	tx1 := chain.GenerateTransaction(aliceWallet, [][]byte{bobAddr}, []int{30})
	chain.AddBlock(bobAddr, "Third Block after genesis", []*transaction.Transaction{tx1})
	// Alice gives Bob 90, David 40, then Bob returns 60, Charlie logs this
	tx2 := chain.GenerateTransaction(aliceWallet, [][]byte{bobAddr, davidAddr}, []int{90, 40})
	tx3 := chain.GenerateTransaction(bobWallet, [][]byte{aliceAddr}, []int{60})
	chain.AddBlock(charlieAddr, "Fourth Block after genesis",
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
	fmt.Printf("Alice: %v.\n", chain.GetBalance(aliceAddr, &aliceWallet.PrivateKey.PublicKey))
	fmt.Printf("Bob: %v.\n", chain.GetBalance(bobAddr, &bobWallet.PrivateKey.PublicKey))
	fmt.Printf("Charlie: %v.\n", chain.GetBalance(charlieAddr, &charlieWallet.PrivateKey.PublicKey))
	fmt.Printf("David: %v.\n", chain.GetBalance(davidAddr, &davidWallet.PrivateKey.PublicKey))
	wallets.SaveFile()
}
