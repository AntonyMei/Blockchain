package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"github.com/AntonyMei/Blockchain/config"
	"github.com/AntonyMei/Blockchain/src/utils"
	"io/ioutil"
	"os"
)

type Wallets struct {
	WalletMap map[string]*Wallet
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
