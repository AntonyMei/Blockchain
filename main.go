package main

import (
	"bufio"
	"fmt"
	"github.com/AntonyMei/Blockchain/config"
	"github.com/AntonyMei/Blockchain/src/blockchain"
	"github.com/AntonyMei/Blockchain/src/cli"
	"github.com/AntonyMei/Blockchain/src/transaction"
	"github.com/AntonyMei/Blockchain/src/utils"
	"github.com/AntonyMei/Blockchain/src/wallet"
	"github.com/AntonyMei/Blockchain/src/network"
	"os"
	"strings"
	"time"
)

func main() {
	run_cli()
}

func ReadCommand(reader *bufio.Reader) []string {
	text, _ := reader.ReadString('\n')
	text = strings.Replace(text, "\n", "", -1)
	text = strings.Replace(text, "\r", "", -1)
	rawInputList := strings.Split(text, " ")
	var inputList []string
	for _, input := range rawInputList {
		if input != "" {
			inputList = append(inputList, input)
		}
	}
	return inputList
}

func run_cli() {
	// login to local system
	fmt.Println("Blockchain interactive mode, type 'Usage' for more information.")
	var reader = bufio.NewReader(os.Stdin)
	var userName string
	var ip string
	var port string
	for {
		fmt.Print(">>> Log in as: ")
		inputList := ReadCommand(reader)
		if len(inputList) == 1 {
			userName = inputList[0]
			pathExists, err := utils.PathExists(config.PersistentStoragePath + userName)
			utils.Handle(err)
			if pathExists {
				fmt.Printf("Login as %v.\n", userName)
			} else {
				fmt.Printf("New user %v.\n", userName)
				err := os.Mkdir(config.PersistentStoragePath+userName, os.ModePerm)
				utils.Handle(err)
			}
			break
		} else {
			fmt.Printf("Expect 1 parameter, got %v instead. To login, enter 'username'\n", len(inputList))
		}
	}
	for {
		fmt.Print(">>> Enter Port: ")
		inputList := ReadCommand(reader)
		if len(inputList) == 1 {
			ip = "localhost" // currently only consider localhost
			port = inputList[0]
			time.Sleep(time.Duration(1) * time.Millisecond)
			break
		} else {
			fmt.Printf("Expect 1 parameter, got %v instead. Enter a valid port number\n", len(inputList))
		}
	}
	walletPath := config.PersistentStoragePath + userName + config.WalletFileName
	blockchainPath := config.PersistentStoragePath + userName + config.BlockchainPath
	fmt.Printf("Wallet path: %v\n", walletPath)
	fmt.Printf("Blockchain path: %v\n", blockchainPath)
	commandLine := cli.InitializeCli(userName, ip, port)

	for {
		// parse input
		fmt.Print(">>>")
		inputList := ReadCommand(reader)

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
		} else if utils.Match(inputList, []string{"ping"}) {
			// ping 
			if !utils.CheckArgumentCount(inputList, 3) {
				continue
			}
			commandLine.Ping(inputList[1], inputList[2])
		} else if utils.Match(inputList, []string{"ls", "connection"}) {
			// ping 
			if !utils.CheckArgumentCount(inputList, 2) {
				continue
			}
			commandLine.CheckConnection()
		} else if utils.Match(inputList, []string{"broadcast"}) {
			// ping 
			if !utils.CheckArgumentCount(inputList, 2) {
				continue
			}
			commandLine.Broadcast(inputList[1])
		} else {
			fmt.Printf("Unknown command.\n")
		}
		fmt.Println()
	}
}

func test_network() {
	println("Network Test")
	agent := os.Args[1]
	ports := map[string]string{
		"Alice": "6000",
		"Bob": "6001",
		"Charlie": "6002",
		"David": "6003",
	}

	pathExists, err := utils.PathExists(config.PersistentStoragePath + agent)
	utils.Handle(err)
	if pathExists {
		fmt.Printf("Login as %v.\n", agent)
	} else {
		fmt.Printf("New user %v.\n", agent)
		err := os.Mkdir(config.PersistentStoragePath+agent, os.ModePerm)
		utils.Handle(err)
	}

	// initialize nodes and wallets for each agent
	var chain *blockchain.BlockChain
	
	wallets, _ := wallet.InitializeWallets(agent)
	// utils.Handle(err)
	// agentAddr := wallets.CreateWallet(agent)
	// agentWallet := wallets.GetWallet(agent)
	if chain == nil {
		chain = blockchain.InitBlockChain(wallets, agent)
	}
	meta := network.NetworkMetaData{Ip:"localhost", Port:ports[agent]}
	// agent_meta := network.UserMetaData{Name:agent, PublicKey: agentWallet.PublicKey, WalletAddr: agentAddr}
	node := network.InitializeNode(wallets, chain, meta)
	node.Serve()
	
	if agent == "Bob" || agent == "Charlie" || agent == "David" {
		alice_meta := network.NetworkMetaData{Ip:"localhost", Port:ports["Alice"]}
		node.SendPingMessage(alice_meta)
	}

	for true {
		time.Sleep(time.Duration(1) * time.Millisecond)
	}
}

func test() {
	println("Wallet Test")
	// initialize wallets
	wallets, err := wallet.InitializeWallets("Alice")
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
	chain := blockchain.InitBlockChain(wallets, "Alice")
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
