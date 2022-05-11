package main

import (
	"bufio"
	"fmt"
	"github.com/AntonyMei/Blockchain/src/blockchain"
	"github.com/AntonyMei/Blockchain/src/cli"
	"github.com/AntonyMei/Blockchain/src/transaction"
	"github.com/AntonyMei/Blockchain/src/utils"
	"github.com/AntonyMei/Blockchain/src/wallet"
	"os"
	"strings"
)

func main() {
	commandLine := cli.InitializeCli()
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Blockchain interactive mode, type 'Usage' for more information.")
	for {
		// parse input
		fmt.Print(">>>")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		rawInputList := strings.Split(text, " ")
		var inputList []string
		for _, input := range rawInputList {
			if input != "" {
				inputList = append(inputList, input)
			}
		}

		// main loop
		if len(inputList) == 0 {
			continue
		}
		if utils.Match(inputList, []string{"exit"}) {
			// exit
			commandLine.Exit()
			return
		} else if utils.Match(inputList, []string{"mk", "wallet"}) {
			// create wallet
			if !utils.CheckArgumentCount(inputList, 3) {
				continue
			}
			commandLine.CreateWallet(inputList[2])
		} else if utils.Match(inputList, []string{"ls", "wallet"}) {
			// check wallet
			if !utils.CheckArgumentCount(inputList, 3) {
				continue
			}
			commandLine.CheckWallet(inputList[2])
		} else if utils.Match(inputList, []string{"ls", "peer"}) {
			// check peer
			if !utils.CheckArgumentCount(inputList, 3) {
				continue
			}
			commandLine.CheckKnownAddress(inputList[2])
		} else {
			fmt.Printf("Unknown command.\n")
		}
		fmt.Println()
	}
}

func test() {
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
		wallets.AddKnownAddress("Alice", &wallet.KnownAddress{Address: aliceAddr,
			PublicKey: aliceWallet.PrivateKey.PublicKey})
		wallets.AddKnownAddress("Bob", &wallet.KnownAddress{Address: bobAddr,
			PublicKey: bobWallet.PrivateKey.PublicKey})
		wallets.AddKnownAddress("Charlie", &wallet.KnownAddress{Address: charlieAddr,
			PublicKey: charlieWallet.PrivateKey.PublicKey})
		wallets.AddKnownAddress("David", &wallet.KnownAddress{Address: davidAddr,
			PublicKey: davidWallet.PrivateKey.PublicKey})
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

	// starts a chain / continues from last chain
	chain := blockchain.InitBlockChain(wallets)
	// alice mines two blocks
	chain.AddBlock(aliceAddr, "Alice 1", []*transaction.Transaction{})
	chain.AddBlock(aliceAddr, "Alice 2", []*transaction.Transaction{})
	// bob comes in and mine another block
	chain.AddBlock(bobAddr, "Bob 1", []*transaction.Transaction{})
	// Alice pays bob 30 in the next block
	tx1 := chain.GenerateTransaction(aliceWallet, [][]byte{bobAddr}, []int{30})
	chain.AddBlock(bobAddr, "Bob records that Alice pays Bob 30.", []*transaction.Transaction{tx1})
	// Alice gives Bob 90, David 40, then Bob returns 60, Charlie logs this
	tx2 := chain.GenerateTransaction(aliceWallet, [][]byte{bobAddr, davidAddr}, []int{90, 40})
	tx3 := chain.GenerateTransaction(bobWallet, [][]byte{aliceAddr}, []int{60})
	chain.AddBlock(charlieAddr, "Charlie records that Alice gives Bob 90, David 40 and Bob returns 60.",
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
	chain.Exit()
}
