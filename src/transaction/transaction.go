package transaction

import "github.com/AntonyMei/Blockchain/config"

type TxRecord struct {
	// Value: number of coins used
	// PubKey: TODO: finish this
	Value  int
	PubKey string
}

type TxSource struct {
	// SourceTxID: ID of source Transaction
	// RecordIdx: index of source TxRecord in source Transaction
	// Sig: TODO: finish this
	SourceTxID []byte
	RecordIdx  int
	Sig        string
}

type Transaction struct {
	TxID         []byte
	SourceList   []TxSource
	TxRecordList []TxRecord
}

func CoinbaseTx(toAddress string, data string) *Transaction {
	// coinbase transaction has no source
	source := TxSource{[]byte{}, -1, data}
	record := TxRecord{config.MiningReward, toAddress}
	transaction := Transaction{nil, []TxSource{source},
		[]TxRecord{record}}
	return &transaction
}
