package transaction

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/AntonyMei/Blockchain/config"
	"github.com/AntonyMei/Blockchain/src/utils"
)

type TxOutput struct {
	// Value: number of coins used
	// PubKey: TODO: finish this
	Value  int
	PubKey string
}

func (txo *TxOutput) CanBeUnlocked(addr string) bool {
	return txo.PubKey == addr
}

func (txo *TxOutput) Log2Terminal() {
	fmt.Printf("[TX Output] Value %v owned by Public Key %v.\n", txo.Value, txo.PubKey)
}

type TxInput struct {
	// SourceTxID: ID of source Transaction
	// TxOutputIdx: index of source TxOutput in source Transaction
	// Sig: TODO: finish this
	SourceTxID  []byte
	TxOutputIdx int
	Sig         string
}

func (source *TxInput) CanUnlock(addr string) bool {
	return source.Sig == addr
}

func (source *TxInput) Log2Terminal() {
	fmt.Printf("[TX Input] Use TXO %v of transaction %x, signed by %v.\n",
		source.TxOutputIdx, source.SourceTxID, source.Sig)
}

type Transaction struct {
	// Note that each address can appear at most once in the output list
	TxID         []byte
	TxInputList  []TxInput
	TxOutputList []TxOutput
}

func (tx *Transaction) SetID() {
	// set TxID as hash value of serialized Transaction
	var encoded bytes.Buffer
	var hash [32]byte
	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	utils.Handle(err)
	hash = sha256.Sum256(encoded.Bytes())
	tx.TxID = hash[:]
}

func (tx *Transaction) IsCoinbase() bool {
	// Check whether a tx is coinbase tx
	return len(tx.TxInputList) == 1 && len(tx.TxInputList[0].SourceTxID) == 0 && tx.TxInputList[0].TxOutputIdx == -1
}

func (tx *Transaction) Log2Terminal() {
	fmt.Printf("[Transaction] TxID %x\n", tx.TxID)
	for _, input := range tx.TxInputList {
		input.Log2Terminal()
	}
	for _, output := range tx.TxOutputList {
		output.Log2Terminal()
	}
	fmt.Println()
}

func CoinbaseTx(minerAddr string, coinbaseSig string) *Transaction {
	// coinbase transaction has no input
	input := TxInput{[]byte{}, -1, coinbaseSig}
	output := TxOutput{config.MiningReward, minerAddr}
	transaction := Transaction{nil, []TxInput{input},
		[]TxOutput{output}}
	return &transaction
}
