package blocks

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/AntonyMei/Blockchain/config"
	"github.com/AntonyMei/Blockchain/src/transaction"
	"github.com/AntonyMei/Blockchain/src/utils"
	"strconv"
)

type Block struct {
	// basic
	PrevHash        []byte
	Hash            []byte
	Data            []byte
	TransactionList []*transaction.Transaction
	// proof of work
	Nonce      int
	Difficulty int
}

func CreateBlock(_data string, txList []*transaction.Transaction, _prevHash []byte, _difficulty int) *Block {
	// create block with given data and difficulty
	newBlock := &Block{PrevHash: _prevHash, Hash: []byte{}, Data: []byte(_data),
		TransactionList: txList, Nonce: 0, Difficulty: _difficulty}
	pow := CreateProofOfWork(newBlock)
	nonce, hash := pow.GenerateNonceHash()
	newBlock.Nonce = nonce
	newBlock.Hash = hash[:]
	return newBlock
}

func Genesis(_difficulty int) *Block {
	// Genesis block is a fixed thing
	input := transaction.TxInput{SourceTxID: []byte{}, TxOutputIdx: -1, Sig: config.CoinbaseSig}
	output := transaction.TxOutput{Value: config.MiningReward, Address: []byte(config.GenesisData)}
	token := make([]byte, 32)
	tx := transaction.Transaction{TxID: token, TxInputList: []transaction.TxInput{input},
		TxOutputList: []transaction.TxOutput{output}}
	tx.SetID()
	return CreateBlock(config.GenesisData, []*transaction.Transaction{&tx}, []byte{}, _difficulty)
}

func (b *Block) Serialize() []byte {
	// serialize a block into byte stream
	var result bytes.Buffer
	var encoder = gob.NewEncoder(&result)
	utils.Handle(encoder.Encode(b))
	return result.Bytes()
}

func Deserialize(stream []byte) *Block {
	// deserialize byte stream from block
	var block Block
	var decoder = gob.NewDecoder(bytes.NewReader(stream))
	utils.Handle(decoder.Decode(&block))
	return &block
}

func (b *Block) GetTransactionsHash() []byte {
	// gather hash value of all transactions, which is their ID
	var txHashList [][]byte
	for _, tx := range b.TransactionList {
		txHashList = append(txHashList, tx.TxID)
	}

	// hash the list to one final hash
	finalHash := sha256.Sum256(bytes.Join(txHashList, []byte{}))
	return finalHash[:]
}

func (b *Block) Log2Terminal() {
	fmt.Printf("****************************************\n")
	fmt.Printf("[Block] %s\n", b.Data)
	fmt.Printf("Hash: %x\n", b.Hash)
	fmt.Printf("Previous Hash: %x\n", b.PrevHash)
	fmt.Printf("Nonce: %v\n", b.Nonce)
	fmt.Printf("Difficulty: %v\n", b.Difficulty)
	pow := CreateProofOfWork(b)
	fmt.Printf("Hash Validated: %s\n", strconv.FormatBool(pow.ValidateNonce()))
	for _, tx := range b.TransactionList {
		tx.Log2Terminal()
	}
	fmt.Printf("****************************************\n\n")
}
