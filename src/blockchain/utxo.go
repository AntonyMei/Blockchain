package blockchain

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"github.com/AntonyMei/Blockchain/config"
	"github.com/AntonyMei/Blockchain/src/utils"
	"io/ioutil"
	"os"
)

type UnspentTXO struct {
	SourceTxID  []byte
	TxOutputIdx int
	Value       int
}

type UTXOSet struct {
	// UTXO Set: address -> (SourceTxID, TxOutputIdx, Value)
	Addr2UTXO   map[string][]UnspentTXO
	UTXOSetPath string
}

func InitUTXOSet(userName string) (*UTXOSet, error) {
	utxoSet := UTXOSet{}
	utxoSet.Addr2UTXO = make(map[string][]UnspentTXO)
	utxoSet.UTXOSetPath = config.PersistentStoragePath + userName + config.UTXOSetPath
	err := utxoSet.LoadFile()
	return &utxoSet, err
}

func (utxoSet *UTXOSet) AddUTXO(addr []byte, txo UnspentTXO) {
	utxoSet.Addr2UTXO[string(addr)] = append(utxoSet.Addr2UTXO[string(addr)], txo)
}

func (utxoSet *UTXOSet) DeleteUTXO(addr []byte, txo UnspentTXO) {
	// Find the UTXO
	var targetIdx = -1
	for idx, utxo := range utxoSet.Addr2UTXO[string(addr)] {
		if bytes.Compare(utxo.SourceTxID, txo.SourceTxID) == 0 && utxo.TxOutputIdx == txo.TxOutputIdx {
			utils.Assert(utxo.Value == txo.Value, "TXO mismatch")
			targetIdx = idx
			break
		}
	}

	// Delete that UTXO
	var length = len(utxoSet.Addr2UTXO[string(addr)])
	utxoSet.Addr2UTXO[string(addr)][targetIdx] = utxoSet.Addr2UTXO[string(addr)][length-1]
	utxoSet.Addr2UTXO[string(addr)] = utxoSet.Addr2UTXO[string(addr)][:length-1]
}

func (utxoSet *UTXOSet) GenerateSpendingPlan(addr []byte, value int) (int, []UnspentTXO) {
	// Generate a spending plan from this UTXOSet
	// if successful: return (total, plan), o.w. return (-1, [])
	var total = 0
	var plan []UnspentTXO
	for _, utxo := range utxoSet.Addr2UTXO[string(addr)] {
		total += utxo.Value
		plan = append(plan, utxo)
		if total >= value {
			break
		}
	}
	if total < value {
		return -1, []UnspentTXO{}
	} else {
		return total, plan
	}
}

func (utxoSet *UTXOSet) SaveFile() {
	// encode the wallets
	var content bytes.Buffer
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(utxoSet)
	utils.Handle(err)
	// save to file
	err = ioutil.WriteFile(utxoSet.UTXOSetPath, content.Bytes(), 0644)
	utils.Handle(err)
}

func (utxoSet *UTXOSet) LoadFile() error {
	// check whether wallet file exists
	if _, err := os.Stat(utxoSet.UTXOSetPath); os.IsNotExist(err) {
		return err
	}

	// read the file
	fileContent, err := ioutil.ReadFile(utxoSet.UTXOSetPath)
	utils.Handle(err)

	// encode it back into a wallet
	var tmpUTXOSet UTXOSet
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&tmpUTXOSet)
	utils.Handle(err)
	utxoSet.Addr2UTXO = tmpUTXOSet.Addr2UTXO
	return nil
}
