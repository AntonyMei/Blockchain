package cli

import (
	"fmt"
	"time"
	"bufio"
	"strings"
	"strconv"
	"encoding/hex"
	"github.com/AntonyMei/Blockchain/src/blockchain"
	"github.com/AntonyMei/Blockchain/src/transaction"
	"github.com/AntonyMei/Blockchain/src/utils"
	"github.com/AntonyMei/Blockchain/src/wallet"
	"github.com/AntonyMei/Blockchain/src/network"
	"github.com/AntonyMei/Blockchain/src/blocks"
	"github.com/AntonyMei/Blockchain/src/blockcache"
)

type Cli struct {
	Wallets      *wallet.Wallets
	Blockchain   *blockchain.BlockChain
	UserName     string
	BlockCache *blockcache.BlockCache
	Node 		 *network.Node
	PendingTxMap *blockchain.PendingTXs
}

// Basic

func InitializeCli(userName string, ip string, port string) *Cli {
	// initialize wallets
	wallets, err := wallet.InitializeWallets(userName)
	if err == nil {
		fmt.Printf("Load wallets succeeded.\n")
	}

	// initialize blockchain
	chain := blockchain.InitBlockChain(wallets, userName)

	// initialize network node
	node := network.InitializeNode(wallets, chain, network.NetworkMetaData{Ip: ip, Port: port})
	node.Serve()

	// initialize cli
	cli := Cli{Wallets: wallets, Blockchain: chain, Node: node}
	cli.BlockCache = blockcache.InitBlockCache(10, chain.LastHash)
	cli.PendingTxMap = blockchain.InitPendingTXs()

	// transaction from network
	node.SetCliTransactionFunc(cli.HandleTxFromNetwork)
	node.SetCliBlockFunc(cli.HandleBlockFromNetwork)
	allBlocks := chain.GetAllBlocks()
	for _, block := range allBlocks {
		node.AddBlock(block)
	}

	// perform some magic op
	transaction.MagicOp()
	return &cli
}

func (commandLine *Cli) Loop(reader *bufio.Reader) {
	s := make(chan string)
	e := make(chan error)

	go func() {
		for true {
			fmt.Print(">>>")
			line, err := reader.ReadString('\n')
			if err != nil {
				e <- err
			} else {
				s <- line
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	tick := time.Tick(100 * time.Millisecond)

MainLoop:
	for {
		select {
		case text := <-s:
			inputList := utils.ParseInput(text)
			// main loop
			if len(inputList) == 0 {
				continue
			}
			if utils.Match(inputList, []string{"exit"}) {
				// exit
				// syntax: exit
				commandLine.Exit()
				return
			} else if utils.Match(inputList, []string{"mk", "wallet"}) {
				// create wallet
				// syntax: mk wallet [name]
				if !utils.CheckArgumentCount(inputList, 3) {
					continue
				}
				commandLine.CreateWallet(inputList[2])
			} else if utils.Match(inputList, []string{"ls", "wallet"}) {
				// list wallet
				// syntax: ls wallet [name/all]
				if !utils.CheckArgumentCount(inputList, 3) {
					continue
				}
				commandLine.ListWallet(inputList[2])
			} else if utils.Match(inputList, []string{"ls", "peer"}) {
				// list peer
				// syntax: ls peer [name/all]
				if !utils.CheckArgumentCount(inputList, 3) {
					continue
				}
				commandLine.ListKnownAddress(inputList[2])
			} else if utils.Match(inputList, []string{"mk", "tx"}) {
				// create new tx
				// syntax: mk tx -n [tx name] -s [sender name] -r [receiver name 1]:[amount 1] ...
				if len(inputList) < 8 || inputList[2] != "-n" || inputList[4] != "-s" || inputList[6] != "-r" {
					fmt.Printf("Syntax error: mk tx -n [tx name] -s [sender name] -r [receiver name 1]:[amount 1] ...\n")
					continue
				}
				txName := inputList[3]
				senderName := inputList[5]
				var receiverNameList []string
				var amountList []int
				for idx := 7; idx < len(inputList); idx++ {
					splitList := strings.Split(inputList[idx], ":")
					if len(splitList) != 2 || len(splitList[0]) == 0 || len(splitList[1]) == 0 {
						fmt.Printf("Syntax error: could not parse receiver list.\n")
						continue MainLoop
					}
					receiverNameList = append(receiverNameList, splitList[0])
					amount, err := strconv.Atoi(splitList[1])
					if err != nil {
						fmt.Printf("Syntax error: could not parse amount.\n")
						continue MainLoop
					}
					amountList = append(amountList, amount)
				}
				commandLine.CreateTransaction(txName, senderName, receiverNameList, amountList)
			} else if utils.Match(inputList, []string{"ls", "tx"}) {
				// list all TXes
				// syntax: ls tx
				commandLine.ListPendingTransactions()
			} else if utils.Match(inputList, []string{"mine"}) {
				// mine a new block
				// syntax: mine -n [miner name] -d [block description] -tx [tx name 1] ...
				if len(inputList) < 5 || inputList[1] != "-n" || inputList[3] != "-d" || len(inputList) == 6 {
					fmt.Printf("Syntax error: mine -n [miner name] -d [block description] -tx [tx name 1] ...\n")
					continue
				}
				minerName := inputList[2]
				blockDescription := inputList[4]
				var txNameList []string
				if len(inputList) > 6 {
					if inputList[5] != "-tx" {
						fmt.Printf("Syntax error: mine -n [miner name] -d [block description] -tx [tx name 1] ...\n")
						continue
					}
					for idx := 6; idx < len(inputList); idx++ {
						txNameList = append(txNameList, inputList[idx])
					}
				}
				go commandLine.MineBlock(minerName, blockDescription, txNameList)
			} else if utils.Match(inputList, []string{"ls", "chain"}) {
				// print the chain
				// syntax: ls chain
				commandLine.PrintBlockchain()
			} else if utils.Match(inputList, []string{"ping"}) {
				// ping 
				if !utils.CheckArgumentCount(inputList, 3) {
					continue
				}
				commandLine.Ping(inputList[1], inputList[2])
			} else if utils.Match(inputList, []string{"ls", "connection"}) {
				// list connections 
				if !utils.CheckArgumentCount(inputList, 2) {
					continue
				}
				commandLine.CheckConnection()
			} else if utils.Match(inputList, []string{"broadcast"}) {
				// broadcast user data
				if !utils.CheckArgumentCount(inputList, 2) {
					continue
				}
				commandLine.Broadcast(inputList[1])
			} else if utils.Match(inputList, []string{"help"}) {
				// print help
				// syntax: help
				commandLine.PrintHelp()
			} else {
				fmt.Printf("Unknown command.\n")
			}
			fmt.Println()
		case <-e:
			continue
		case <-time.After(time.Duration(10) * time.Millisecond):
			// handle blocks from network
			commandLine.HandleBlock()

			// ping a random node to catch up chain
			commandLine.Node.RandomPing(commandLine.Blockchain.BlockHeight)
		case <-tick:
			// broadcast all private users' id
			accountNames := commandLine.Wallets.GetAllWalletNames()
			for _, name := range accountNames {
				wallet := commandLine.Wallets.GetWallet(name)
				user_meta := network.UserMetaData{Name: name, PublicKey: wallet.PublicKey, WalletAddr: wallet.Address()}
				commandLine.Node.BroadcastUserMessage(user_meta)
			}
		}
	}
}

func (cli *Cli) Exit() {
	cli.Wallets.SaveFile()
	cli.Blockchain.Exit()
}

// Wallets

func (cli *Cli) CreateWallet(name string) {
	if name == "All" || name == "all" {
		fmt.Printf("All / all is reserved name.\n")
		return
	}
	if tmp := cli.Wallets.GetWallet(name); tmp != nil {
		fmt.Printf("Wallet with name %s already exists.\n", name)
		return
	}
	addr := cli.Wallets.CreateWallet(name)
	fmt.Printf("Wallet: %s\n", name)
	fmt.Printf("Address: %x\n", addr)
	// put this address into known addresses
	res := cli.Wallets.GetWallet(name)
	cli.Wallets.AddKnownAddress(name, &wallet.KnownAddress{Address: addr, PublicKey: res.PrivateKey.PublicKey})
}

func (cli *Cli) ListWallet(name string) {
	if name == "All" || name == "all" {
		cli._listAllWallets()
	} else {
		cli._listWallet(name)
	}
}

func (cli *Cli) _listWallet(name string) {
	res := cli.Wallets.GetWallet(name)
	if res == nil {
		fmt.Printf("Error: no wallet with name %s.\n", name)
		return
	}
	addr := res.Address()
	fmt.Printf("Wallet: %s\n", name)
	fmt.Printf("Address: %x\n", addr)
	balance := cli.Blockchain.GetBalance(addr, &res.PrivateKey.PublicKey)
	fmt.Printf("Balance: %v\n", balance)
}

func (cli *Cli) _listAllWallets() {
	var accountNames = cli.Wallets.GetAllWalletNames()
	for _, name := range accountNames {
		cli._listWallet(name)
		fmt.Println()
	}
}

func (cli *Cli) ListKnownAddress(name string) {
	if name == "All" || name == "all" {
		cli._listAllKnownAddresses()
	} else {
		cli._listKnownAddress(name)
	}
}

func (cli *Cli) _listKnownAddress(name string) {
	res := cli.Wallets.GetKnownAddress(name)
	if res == nil {
		fmt.Printf("Error: no known address with name %s.\n", name)
		return
	}
	fmt.Printf("Known Address: %s has address %x.\n", name, res.Address)
}

func (cli *Cli) _listAllKnownAddresses() {
	for name := range cli.Wallets.KnownAddressMap {
		cli._listKnownAddress(name)
	}
}

func (cli *Cli) CreateTransaction(txName string, sender string, receiverList []string, amountList []int) string {
	// check input shape
	if len(receiverList) != len(amountList) {
		fmt.Printf("Error: receiver list and amount list shape mismatch.\n")
	}

	// get sender wallet and receiver addresses
	fromWallet := cli.Wallets.GetWallet(sender)
	if fromWallet == nil {
		fmt.Printf("Error: No wallet with name %s.\n", sender)
		return ""
	}
	var toAddrList [][]byte
	for _, receiver := range receiverList {
		receiverAddr := cli.Wallets.GetKnownAddress(receiver)
		if receiverAddr == nil {
			fmt.Printf("Error: No known address with name %s.\n", receiver)
			return ""
		}
		toAddrList = append(toAddrList, receiverAddr.Address)
	}

	// create TX and put into pending zone
	newTX := cli.Blockchain.GenerateTransaction(fromWallet, toAddrList, amountList)
	txKey := txName + "::" + string(utils.Base58Encode(newTX.TxID[:8]))
	cli.PendingTxMap.AddTransaction(txKey, newTX)
	fmt.Printf("New transaction: %s.\n", txKey)

	// broadcast transaction
	cli.Node.BroadcastTransaction(txKey, newTX)
	return txKey
}

func (cli *Cli) ListPendingTransactions() {
	cli.PendingTxMap.ListPendingTransactions()
}

func (cli *Cli) MineBlock(minerName string, description string, txNameList []string) {
	// get miner wallet
	minerWallet := cli.Wallets.GetWallet(minerName)
	if minerWallet == nil {
		fmt.Printf("Error: No wallet with name %s.\n", minerName)
		return
	}
	// get tx and remove from pending tx list
	var blockTXList []*transaction.Transaction
	for _, txName := range txNameList {
		tx := cli.PendingTxMap.GetTx(txName)
		if tx == nil {
			fmt.Printf("Error: no transaction with name %s.\n", txName)
			return
		}
		blockTXList = append(blockTXList, tx)
	}
	// mine a new block
	newBlock := cli.Blockchain.MineBlock(minerWallet.Address(), description, blockTXList)
	// put the block into the cache
	cli.BlockCache.AddBlock(newBlock)
}

func (cli *Cli) PrintBlockchain() {
	cli.Blockchain.Log2Terminal()
}

// Network

func (cli *Cli) Ping(ip string, port string) {
	cli.Node.SendPingMessage(network.NetworkMetaData{Ip: ip, Port: port}, cli.Blockchain.BlockHeight)
}

func (cli *Cli) CheckConnection() {
	cli.Node.ConnectionPool.ShowPool()
}

func (cli *Cli) Broadcast(name string) {
	wallet := cli.Wallets.GetWallet(name)
	if wallet == nil {
		fmt.Printf("Error: no wallet with name %s.\n", name)
		return
	}
	user_meta := network.UserMetaData{Name: name, PublicKey: wallet.PublicKey, WalletAddr: wallet.Address()}
	cli.Node.BroadcastUserMessage(user_meta)
}

func (cli *Cli) HandleTxFromNetwork(txKey string, tx *transaction.Transaction) {
	cur_tx := cli.PendingTxMap.GetTx(txKey)
	if cur_tx == nil {
		cli.PendingTxMap.AddTransaction(txKey, tx)
		// fmt.Printf("Receive transaction from network: %s.\n", txKey)

		// broadcast again
		cli.Node.BroadcastTransaction(txKey, tx)
	}
}

func (cli *Cli) HandleBlockFromNetwork(block *blocks.Block) {
	// puts the block into a cache
	cli.BlockCache.AddBlock(block)
}

func (cli *Cli) HandleBlock() {
	// handle block from cache
	block := cli.BlockCache.PopBlock()
	if block != nil {
		validBlock := cli.Blockchain.AddBlock(block)
		if validBlock {
			cli.BlockCache.SetLastHash(cli.Blockchain.LastHash)
			cli.RemoveMinedTXs(block)
			cli.Node.AddBlock(block)
			cli.Node.BroadcastBlockSource(block)
		}
	}
}

func (cli *Cli) RemoveMinedTXs(block *blocks.Block) {
	// remove duplicate pending transactions
	allPendingTxKeys, allPendingTxs := cli.PendingTxMap.GetAllTx()
	MinedTxs := make(map[string]bool)
	for _, tx := range block.TransactionList {
		MinedTxs[hex.EncodeToString(tx.TxID)] = true
	}
	for i, tx := range allPendingTxs {
		txKey := allPendingTxKeys[i]
		_, exists := MinedTxs[hex.EncodeToString(tx.TxID)]
		if exists {
			cli.PendingTxMap.DeleteTx(txKey)
		}
	}
}

func (cli *Cli) PrintHelp() {
	fmt.Println("[1] print help              help")
	fmt.Println("[2] create wallet           mk wallet [name]")
	fmt.Println("    create new TX           mk tx -n [tx name] -s [sender name] -r [receiver name 1]:[amount 1] ...")
	fmt.Println("    mine a new block        mine -n [miner name] -d [block description] -tx [tx name 1] ...")
	fmt.Println("[3] list wallet             ls wallet [name/all]")
	fmt.Println("    list peer syntax        ls peer [name/all]")
	fmt.Println("    list all pending TXes   ls tx")
	fmt.Println("    print whole chain       ls chain")
	fmt.Println("[4] ping a node             ping [ip] [port]")
	fmt.Println("    broadcast user name     broadcast [user name]")
	fmt.Println("    list known nodes        ls connection")
	fmt.Println("[5] exit                    exit")
}
