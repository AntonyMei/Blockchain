package blockchain

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"encoding/hex"
	"github.com/AntonyMei/Blockchain/config"
	"github.com/AntonyMei/Blockchain/src/blocks"
	"github.com/AntonyMei/Blockchain/src/utils"
	"io/ioutil"
	"os"
	"strconv"
)

type UnspentTXO struct {
	SourceTxID  []byte
	TxOutputIdx int
	Value       int
}

type UTXOSet struct {
	// UTXO Set: address -> (SourceTxID, TxOutputIdx, Value)
	Addr2UTXO   map[string][]UnspentTXO
	UTXO2Addr   map[string]string
	UTXOSetPath string
}

func InitUTXOSet(userName string) (*UTXOSet, error) {
	utxoSet := UTXOSet{}
	utxoSet.Addr2UTXO = make(map[string][]UnspentTXO)
	utxoSet.UTXO2Addr = make(map[string]string)
	utxoSet.UTXOSetPath = config.PersistentStoragePath + userName + config.UTXOSetPath
	err := utxoSet.LoadFile()
	return &utxoSet, err
}

func (utxoSet *UTXOSet) AddUTXO(addr []byte, txo UnspentTXO) {
	// save utxo into addr -> utxo map
	utxoSet.Addr2UTXO[string(addr)] = append(utxoSet.Addr2UTXO[string(addr)], txo)
	// save utxo into (TxID | OutputIdx) -> addr map
	UTXOKey := string(txo.SourceTxID) + strconv.Itoa(txo.TxOutputIdx)
	utxoSet.UTXO2Addr[UTXOKey] = string(addr)
}

func (utxoSet *UTXOSet) DeleteUTXO(addr []byte, txo UnspentTXO) {
	// find the UTXO in addr -> utxo map
	var targetIdx = -1
	for idx, utxo := range utxoSet.Addr2UTXO[string(addr)] {
		if bytes.Compare(utxo.SourceTxID, txo.SourceTxID) == 0 && utxo.TxOutputIdx == txo.TxOutputIdx {
			targetIdx = idx
			break
		}
	}

	// delete that UTXO from addr -> utxo map
	var length = len(utxoSet.Addr2UTXO[string(addr)])
	utxoSet.Addr2UTXO[string(addr)][targetIdx] = utxoSet.Addr2UTXO[string(addr)][length-1]
	utxoSet.Addr2UTXO[string(addr)] = utxoSet.Addr2UTXO[string(addr)][:length-1]

	// delete that UTXO from (TxID | OutputIdx) -> addr map
	UTXOKey := string(txo.SourceTxID) + strconv.Itoa(txo.TxOutputIdx)
	delete(utxoSet.UTXO2Addr, UTXOKey)
}

func (utxoSet *UTXOSet) DumpBlock(block *blocks.Block) {
	// Put every output of the given block into UTXO set and remove its inputs
	for _, tx := range block.TransactionList {
		// remove input
		for _, input := range tx.TxInputList {
			UTXOKey := string(input.SourceTxID) + strconv.Itoa(input.TxOutputIdx)
			addr := utxoSet.UTXO2Addr[UTXOKey]
			utxoSet.DeleteUTXO([]byte(addr), UnspentTXO{
				SourceTxID:  input.SourceTxID,
				TxOutputIdx: input.TxOutputIdx,
				Value:       -1,
			})
		}
		// dump output
		for idx, txo := range tx.TxOutputList {
			utxoSet.AddUTXO(txo.Address, UnspentTXO{
				SourceTxID:  tx.TxID,
				TxOutputIdx: idx,
				Value:       txo.Value,
			})
		}
	}
}

func (utxoSet *UTXOSet) GenerateSpendingPlan(addr []byte, value int) (int, map[string][]int) {
	var total, unspentList = utxoSet._GenerateSpendingPlan(addr, value)
	if total != value {
		return total, make(map[string][]int)
	} else {
		var candidateList = make(map[string][]int)
		for _, utxo := range unspentList {
			txID := hex.EncodeToString(utxo.SourceTxID)
			candidateList[txID] = append(candidateList[txID], utxo.TxOutputIdx)
		}
		return total, candidateList
	}
}

func (utxoSet *UTXOSet) _GenerateSpendingPlan(addr []byte, value int) (int, []UnspentTXO) {
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
	utxoSet.UTXO2Addr = tmpUTXOSet.UTXO2Addr
	return nil
}
