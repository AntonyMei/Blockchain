package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"github.com/AntonyMei/Blockchain/config"
	"github.com/AntonyMei/Blockchain/src/utils"
	"io/ioutil"
	"os"
)

type Wallets struct {
	WalletMap map[string]*Wallet
}

func InitializeWallets() (*Wallets, error) {
	// create new wallets
	wallets := Wallets{}
	wallets.WalletMap = make(map[string]*Wallet)
	err := wallets.LoadFile()
	return &wallets, err
}

func (ws *Wallets) CreateWallet(name string) string {
	// returns address of that wallet
	wallet := CreateWallet()
	ws.WalletMap[name] = wallet
	address := fmt.Sprintf("%s", wallet.Address())
	return address
}

func (ws Wallets) GetWallet(name string) Wallet {
	return *ws.WalletMap[name]
}

func (ws *Wallets) GetAllAccounts() []string {
	var accountNames []string
	for name := range ws.WalletMap {
		accountNames = append(accountNames, name)
	}
	return accountNames
}

func (ws *Wallets) SaveFile() {
	// encode the wallets
	var content bytes.Buffer
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	utils.Handle(err)
	// save to file
	err = ioutil.WriteFile(config.WalletPath, content.Bytes(), 0644)
	utils.Handle(err)
}

func (ws *Wallets) LoadFile() error {
	// check whether wallet file exists
	if _, err := os.Stat(config.WalletPath); os.IsNotExist(err) {
		return err
	}

	// read the file
	fileContent, err := ioutil.ReadFile(config.WalletPath)
	utils.Handle(err)

	// encode it back into a wallet
	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	utils.Handle(err)
	ws.WalletMap = wallets.WalletMap
	return nil
}
