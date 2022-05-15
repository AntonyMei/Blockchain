package wallet

import (
	"bytes"
	"sync"
	"crypto/elliptic"
	"encoding/gob"
	"github.com/AntonyMei/Blockchain/config"
	"github.com/AntonyMei/Blockchain/src/utils"
	"io/ioutil"
	"os"
)

type Wallets struct {
	PersonalWalletMap map[string]*Wallet
	KnownAddressMap   map[string]*KnownAddress
	WalletPath        string
	mu                sync.Mutex
}

func InitializeWallets(userName string) (*Wallets, error) {
	// create new wallets
	wallets := Wallets{}
	wallets.PersonalWalletMap = make(map[string]*Wallet)
	wallets.KnownAddressMap = make(map[string]*KnownAddress)
	wallets.WalletPath = config.PersistentStoragePath + userName + config.WalletFileName
	err := wallets.LoadFile()
	return &wallets, err
}

func (ws *Wallets) CreateWallet(name string) []byte {
	// returns address of that wallet
	wallet := CreateWallet()
	ws.PersonalWalletMap[name] = wallet
	return wallet.Address()
}

func (ws *Wallets) AddWallet(name string, wallet *Wallet) {
	ws.PersonalWalletMap[name] = wallet
}

func (ws *Wallets) AddKnownAddress(name string, knownAddress *KnownAddress) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.KnownAddressMap[name] = knownAddress
}

func (ws *Wallets) GetWallet(name string) *Wallet {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.PersonalWalletMap[name]
}

func (ws *Wallets) GetKnownAddress(name string) *KnownAddress {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.KnownAddressMap[name]
}

func (ws *Wallets) GetAllWalletNames() []string {
	var accountNames []string
	for name := range ws.PersonalWalletMap {
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
	err = ioutil.WriteFile(ws.WalletPath, content.Bytes(), 0644)
	utils.Handle(err)
}

func (ws *Wallets) LoadFile() error {
	// check whether wallet file exists
	if _, err := os.Stat(ws.WalletPath); os.IsNotExist(err) {
		return err
	}

	// read the file
	fileContent, err := ioutil.ReadFile(ws.WalletPath)
	utils.Handle(err)

	// encode it back into a wallet
	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	utils.Handle(err)
	ws.PersonalWalletMap = wallets.PersonalWalletMap
	ws.KnownAddressMap = wallets.KnownAddressMap
	return nil
}
