package cli

import (
	"fmt"
	"github.com/AntonyMei/Blockchain/src/blockchain"
	"github.com/AntonyMei/Blockchain/src/transaction"
	"github.com/AntonyMei/Blockchain/src/wallet"
	"github.com/AntonyMei/Blockchain/src/network"
)

type Cli struct {
	Wallets    *wallet.Wallets
	Blockchain *blockchain.BlockChain
	Node 	   *network.Node
	UserName   string
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

func (cli *Cli) CheckWallet(name string) {
	if name == "All" || name == "all" {
		cli._checkAllWallets()
	} else {
		cli._checkWallet(name)
	}
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

func (cli *Cli) _checkWallet(name string) {
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

func (cli *Cli) _checkAllWallets() {
	var accountNames = cli.Wallets.GetAllWalletNames()
	for _, name := range accountNames {
		cli._checkWallet(name)
		fmt.Println()
	}
}

func (cli *Cli) CheckKnownAddress(name string) {
	if name == "All" || name == "all" {
		cli._checkAllKnownAddresses()
	} else {
		cli._checkKnownAddress(name)
	}
}

func (cli *Cli) _checkKnownAddress(name string) {
	res := cli.Wallets.GetKnownAddress(name)
	if res == nil {
		fmt.Printf("Error: no known address with name %s.\n", name)
		return
	}
	fmt.Printf("Known Address: %s has address %x.\n", name, res.Address)
}

func (cli *Cli) _checkAllKnownAddresses() {
	for name := range cli.Wallets.KnownAddressMap {
		cli._checkKnownAddress(name)
	}
}
