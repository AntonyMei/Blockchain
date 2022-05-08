package transaction

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"github.com/AntonyMei/Blockchain/config"
	"github.com/AntonyMei/Blockchain/src/utils"
)

type TxOutput struct {
	// Value: number of coins used
	// PubKey: TODO: finish this
	Value  int
	PubKey string
}

func (txo *TxOutput) CanUnlock(data string) bool {
	return txo.PubKey == data
}

type TxInput struct {
	// SourceTxID: ID of source Transaction
	// TxOutputIdx: index of source TxOutput in source Transaction
	// Sig: TODO: finish this
	SourceTxID  []byte
	TxOutputIdx int
	Sig         string
}

func (source *TxInput) CanUnlock(data string) bool {
	return source.Sig == data
}

type Transaction struct {
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

func CoinbaseTx(minerAddr string, data string) *Transaction {
	// coinbase transaction has no input
	input := TxInput{[]byte{}, -1, data}
	output := TxOutput{config.MiningReward, minerAddr}
	transaction := Transaction{nil, []TxInput{input},
		[]TxOutput{output}}
	return &transaction
}
