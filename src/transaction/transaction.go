package transaction

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/AntonyMei/Blockchain/config"
	"github.com/AntonyMei/Blockchain/src/utils"
)

type TxOutput struct {
	// Value: number of coins used
	// Address: address of receiver
	Value   int
	Address []byte
}

func (txo *TxOutput) BelongsTo(addr []byte) bool {
	return bytes.Compare(txo.Address, addr) == 0
}

func (txo *TxOutput) Log2Terminal() {
	fmt.Printf("[TX Output] Give %v coins to account %x.\n", txo.Value, txo.Address)
}

func (txo *TxOutput) Serialize() []byte {
	serialized := bytes.Join([][]byte{txo.Address, utils.Int2Hex(int64(txo.Value))}, []byte{})
	return serialized
}

type TxInput struct {
	// SourceTxID: ID of source Transaction
	// TxOutputIdx: index of source TxOutput in source Transaction
	// Sig: signed by owner of source TXO
	SourceTxID  []byte
	TxOutputIdx int
	Sig         string
}

func (source *TxInput) Sign(privateKey *ecdsa.PrivateKey) {
	// hash the TxInput into a byte array
	source.Sig = ""
	var result bytes.Buffer
	var encoder = gob.NewEncoder(&result)
	utils.Handle(encoder.Encode(source))
	hashedValue := sha256.Sum256(result.Bytes())
	// sign the hashed value with privateKey
	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, hashedValue[:])
	utils.Handle(err)
	source.Sig = string(signature)
}

func (source *TxInput) Verify(publicKey *ecdsa.PublicKey) bool {
	// if input value is nil, we check whether it is coinbase signature
	if publicKey == nil {
		return source.Sig == config.CoinbaseSig
	}
	// hash a Copy of source into a byte array
	sourceCopy := TxInput{SourceTxID: source.SourceTxID, TxOutputIdx: source.TxOutputIdx, Sig: ""}
	var stream bytes.Buffer
	var encoder = gob.NewEncoder(&stream)
	utils.Handle(encoder.Encode(sourceCopy))
	hashedValue := sha256.Sum256(stream.Bytes())
	// check whether the signature is correct
	result := ecdsa.VerifyASN1(publicKey, hashedValue[:], []byte(source.Sig))
	return result
}

func (source *TxInput) Log2Terminal() {
	fmt.Printf("[TX Input] Use TXO %v of transaction %x.\n",
		source.TxOutputIdx, source.SourceTxID)
}

func (source *TxInput) Serialize() []byte {
	serialized := bytes.Join([][]byte{source.SourceTxID, utils.Int2Hex(int64(source.TxOutputIdx)),
		[]byte(source.Sig)}, []byte{})
	return serialized
}

type Transaction struct {
	// Note that each address can appear at most once in the output list
	TxID         []byte
	TxInputList  []TxInput
	TxOutputList []TxOutput
	//Str            string // a meaningless string only for making the transaction large (test network bytes use)
}

func (tx *Transaction) SetID() {
	// set TxID as hash value of serialized Transaction
	// serialize tx into byte stream
	/* turn on when testing network data transfer
	b := make([]rune, 10000)
	for i := range b {
		b[i] = rune('a')
	}
	tx.Str = string(b)*/
	//tx.Str = ""
	var raw []byte
	for _, input := range tx.TxInputList {
		raw = bytes.Join([][]byte{raw, input.Serialize()}, []byte{})
	}
	for _, output := range tx.TxOutputList {
		raw = bytes.Join([][]byte{raw, output.Serialize()}, []byte{})
	}
	// set TxID
	var hash [32]byte
	hash = sha256.Sum256(raw)
	tx.TxID = hash[:]
	//tx.Log2Terminal()
}

func (tx *Transaction) IsCoinbase() bool {
	// Check whether a tx is coinbase tx
	condition1 := len(tx.TxInputList) == 1 && len(tx.TxInputList[0].SourceTxID) == 0 && tx.TxInputList[0].TxOutputIdx == -1 && tx.TxInputList[0].Sig == config.CoinbaseSig
	condition2 := len(tx.TxOutputList) == 1 && tx.TxOutputList[0].Value == config.MiningReward
	return condition1 && condition2
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

func CoinbaseTx(minerAddr []byte) *Transaction {
	// coinbase transaction has no input, and gives MiningReward to miner
	input := TxInput{[]byte{}, -1, config.CoinbaseSig}
	output := TxOutput{config.MiningReward, minerAddr}
	// to identify different coinbase TXes, we add randomness to initial TxID
	token := make([]byte, 32)
	_, _ = rand.Read(token)
	transaction := Transaction{token, []TxInput{input}, []TxOutput{output}}
	transaction.SetID()
	return &transaction
}

func MagicOp() {
	transactions := TxOutput{}
	var encoded bytes.Buffer
	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(transactions)
	utils.Handle(err)
}
