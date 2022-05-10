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
	PersonalWallets map[string]*Wallet
	KnownAddresses  map[string][]byte
}

func InitializeWallets() (*Wallets, error) {
	// create new wallets
	wallets := Wallets{}
	wallets.PersonalWallets = make(map[string]*Wallet)
	err := wallets.LoadFile()
	return &wallets, err
}

func (ws *Wallets) CreateWallet(name string) string {
	// returns address of that wallet
	wallet := CreateWallet()
	ws.PersonalWallets[name] = wallet
	address := fmt.Sprintf("%s", wallet.Address())
	return address
}

func (ws Wallets) GetPersonalWallet(name string) Wallet {
	return *ws.PersonalWallets[name]
}

func (ws Wallets) GetKnownAddress(name string) []byte {
	return ws.KnownAddresses[name]
}

func (ws *Wallets) GetAllPersonalWalletNames() []string {
	var accountNames []string
	for name := range ws.PersonalWallets {
		accountNames = append(accountNames, name)
	}
	return accountNames
}

func (ws *Wallets) GetAllKnownAddresses() map[string][]byte {
	return ws.KnownAddresses
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
	ws.PersonalWallets = wallets.PersonalWallets
	return nil
}
