package cli

import (
	"fmt"
	"github.com/AntonyMei/Blockchain/src/blockchain"
	"github.com/AntonyMei/Blockchain/src/transaction"
	"github.com/AntonyMei/Blockchain/src/utils"
	"github.com/AntonyMei/Blockchain/src/wallet"
	"github.com/AntonyMei/Blockchain/src/network"
)

type Cli struct {
	Wallets      *wallet.Wallets
	Blockchain   *blockchain.BlockChain
	UserName     string
	pendingTXMap *blockchain.PendingTXs
	Node 		 *network.Node
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
	cli.pendingTXMap = blockchain.InitPendingTXs()

	// transaction from network
	node.SetCliTransactionFunc(cli.HandleTxFromNetwork)

	// perform some magic op
	transaction.MagicOp()
	return &cli
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

func (cli *Cli) CreateTransaction(txName string, sender string, receiverList []string, amountList []int) {
	// check input shape
	if len(receiverList) != len(amountList) {
		fmt.Printf("Error: receiver list and amount list shape mismatch.\n")
	}

	// get sender wallet and receiver addresses
	fromWallet := cli.Wallets.GetWallet(sender)
	if fromWallet == nil {
		fmt.Printf("Error: No wallet with name %s.\n", sender)
		return
	}
	var toAddrList [][]byte
	for _, receiver := range receiverList {
		receiverAddr := cli.Wallets.GetKnownAddress(receiver)
		if receiverAddr == nil {
			fmt.Printf("Error: No known address with name %s.\n", receiver)
			return
		}
		toAddrList = append(toAddrList, receiverAddr.Address)
	}

	// create TX and put into pending zone
	newTX := cli.Blockchain.GenerateTransaction(fromWallet, toAddrList, amountList)
	txKey := txName + "::" + string(utils.Base58Encode(newTX.TxID[:8]))
	cli.pendingTXMap.AddTransaction(txKey, newTX)
	fmt.Printf("New transaction: %s.\n", txKey)

	// broadcast transaction
	cli.Node.BroadcastTransaction(txKey, newTX)
}

func (cli *Cli) ListPendingTransactions() {
	cli.pendingTXMap.ListPendingTransactions()
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
		tx := cli.pendingTXMap.GetTx(txName)
		if tx == nil {
			fmt.Printf("Error: no transaction with name %s.\n", txName)
			return
		}
		blockTXList = append(blockTXList, tx)
	}
	for _, txName := range txNameList {
		cli.pendingTXMap.DeleteTx(txName)
	}
	// add a new block
	cli.Blockchain.AddBlock(minerWallet.Address(), description, blockTXList)
}

func (cli *Cli) PrintBlockchain() {
	cli.Blockchain.Log2Terminal()
}

// Network

func (cli *Cli) Ping(ip string, port string) {
	cli.Node.SendPingMessage(network.NetworkMetaData{Ip: ip, Port: port})
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
	cur_tx := cli.pendingTXMap.GetTx(txKey)
	if cur_tx == nil {
		cli.pendingTXMap.AddTransaction(txKey, tx)
		fmt.Printf("Receive transaction from network: %s.\n", txKey)

		// broadcast again
		cli.Node.BroadcastTransaction(txKey, tx)
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
	fmt.Println("[4] exit                    exit")
}
