package cli

import (
	"fmt"
	"github.com/AntonyMei/Blockchain/src/blockchain"
	"github.com/AntonyMei/Blockchain/src/transaction"
	"github.com/AntonyMei/Blockchain/src/wallet"
)

type Cli struct {
	Wallets      *wallet.Wallets
	Blockchain   *blockchain.BlockChain
	UserName     string
	pendingTXMap map[string]*transaction.Transaction
}

// Basic

func InitializeCli(userName string) *Cli {
	// initialize wallets
	wallets, err := wallet.InitializeWallets(userName)
	if err == nil {
		fmt.Printf("Load wallets succeeded.\n")
	}

	// initialize blockchain
	chain := blockchain.InitBlockChain(wallets, userName)

	// initialize cli
	cli := Cli{Wallets: wallets, Blockchain: chain}

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
		fmt.Printf("Error: No wallet with name %x.\n", sender)
		return
	}
	var toAddrList [][]byte
	for _, receiver := range receiverList {
		receiverAddr := cli.Wallets.GetKnownAddress(receiver)
		if receiverAddr == nil {
			fmt.Printf("Error: No known address with name %x.\n", receiver)
			return
		}
		toAddrList = append(toAddrList, receiverAddr.Address)
	}

	// create TX and put into pending zone
	newTX := cli.Blockchain.GenerateTransaction(fromWallet, toAddrList, amountList)
	txKey := txName + "::" + string(newTX.TxID[:8])
	cli.pendingTXMap[txKey] = newTX
	fmt.Printf("New transaction: %x.\n", txKey)
}

func (cli *Cli) ListPendingTransactions() {
	idx := 0
	for txKey := range cli.pendingTXMap {
		fmt.Printf("Transaction %v: %x\n", idx, txKey)
		idx += 1
	}
}
