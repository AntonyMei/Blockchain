package transaction

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"github.com/AntonyMei/Blockchain/config"
	"github.com/AntonyMei/Blockchain/src/utils"
)

type TxRecord struct {
	// Value: number of coins used
	// PubKey: TODO: finish this
	Value  int
	PubKey string
}

func (record *TxRecord) CanUnlock(data string) bool {
	return record.PubKey == data
}

type TxSource struct {
	// SourceTxID: ID of source Transaction
	// RecordIdx: index of source TxRecord in source Transaction
	// Sig: TODO: finish this
	SourceTxID []byte
	RecordIdx  int
	Sig        string
}

func (source *TxSource) CanUnlock(data string) bool {
	return source.Sig == data
}

type Transaction struct {
	TxID         []byte
	SourceList   []TxSource
	TxRecordList []TxRecord
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
	return len(tx.SourceList) == 1 && len(tx.SourceList[0].SourceTxID) == 0 && tx.SourceList[0].RecordIdx == -1
}

func CoinbaseTx(toAddress string, data string) *Transaction {
	// coinbase transaction has no source
	source := TxSource{[]byte{}, -1, data}
	record := TxRecord{config.MiningReward, toAddress}
	transaction := Transaction{nil, []TxSource{source},
		[]TxRecord{record}}
	return &transaction
}
