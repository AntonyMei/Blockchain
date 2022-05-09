package blockchain

import (
	"encoding/hex"
	"fmt"
	"github.com/AntonyMei/Blockchain/config"
	"github.com/AntonyMei/Blockchain/src/blocks"
	"github.com/AntonyMei/Blockchain/src/transaction"
	"github.com/AntonyMei/Blockchain/src/utils"
	"github.com/dgraph-io/badger"
	"log"
)

type BlockChain struct {
	// blockchain is stored in badger database (k-v database)
	// key: hash of block, value: serialized block
	Database *badger.DB
	// proof of difficulty
	ChainDifficulty int
}

func InitBlockChain(genesisMinerAddr string) *BlockChain {
	// open db connection
	var options = badger.DefaultOptions(config.PersistentStoragePath)
	database, err := badger.Open(options)
	utils.Handle(err)

	// create a new blockchain if nothing exists
	err = database.Update(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("lasthash"))
		if err == badger.ErrKeyNotFound {
			// no chain in database, create a new one
			fmt.Println("Initiating a new blockchain...")
			coinbaseTx := transaction.CoinbaseTx(genesisMinerAddr, config.CoinbaseSig)
			genesis := blocks.Genesis(coinbaseTx, config.InitialChainDifficulty)
			err = txn.Set(genesis.Hash, genesis.Serialize())
			utils.Handle(err)
			err = txn.Set([]byte("lasthash"), genesis.Hash)
			utils.Handle(err)
		} else {
			// there exists a blockchain already
			fmt.Println("Continuing from saved blockchain...")
		}
		return nil
	})
	utils.Handle(err)

	blockchain := BlockChain{Database: database, ChainDifficulty: config.InitialChainDifficulty}
	return &blockchain
}

func (bc *BlockChain) AddBlock(minerAddr string, description string, txList []*transaction.Transaction) {
	// add block should be a database transaction
	err := bc.Database.Update(func(txn *badger.Txn) error {
		// get last hash from database
		var lastHash []byte
		item, err := txn.Get([]byte("lasthash"))
		utils.Handle(err)
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		utils.Handle(err)

		// create new block and write into db
		txList = append(txList, transaction.CoinbaseTx(minerAddr, config.CoinbaseSig))
		newBlock := blocks.CreateBlock(description, txList, lastHash, bc.ChainDifficulty)
		err = txn.Set(newBlock.Hash, newBlock.Serialize())
		utils.Handle(err)
		err = txn.Set([]byte("lasthash"), newBlock.Hash)
		utils.Handle(err)
		return nil
	})
	utils.Handle(err)
}

func (bc *BlockChain) FindUnspentTransactions(address string) []transaction.Transaction {
	// This function returns all transactions that contain unspent outputs associated with address

	// initialize
	var unspentTxs []transaction.Transaction
	spentTxMap := make(map[string][]int)
	bcIterator := bc.Iterator()

	// iterator through the chain to find unspent transactions
	for {
		// read block from chain
		block := bcIterator.GetVal()
		bcIterator.Next()

		// check each transaction in the list
		for _, tx := range block.TransactionList {
			txID := hex.EncodeToString(tx.TxID)

		Outputs:
			// check each TxOutput
			for outIdx, out := range tx.TxOutputList {
				if spentTxMap[txID] != nil {
					for _, spentOutIdx := range spentTxMap[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}
				if out.CanBeUnlocked(address) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			// mark all its inputs as spent
			if tx.IsCoinbase() == false {
				for _, in := range tx.TxInputList {
					if in.CanUnlock(address) {
						inTxID := hex.EncodeToString(in.SourceTxID)
						spentTxMap[inTxID] = append(spentTxMap[inTxID], in.TxOutputIdx)
					}
				}
			}
			if len(block.PrevHash) == 0 {
				break
			}
		}
		return unspentTxs
	}
}

func (bc *BlockChain) FindUTXO(address string) []transaction.TxOutput {
	// This function returns all UTXOs associated with address
	var UTXOs []transaction.TxOutput
	unspentTransactions := bc.FindUnspentTransactions(address)
	for _, tx := range unspentTransactions {
		for _, out := range tx.TxOutputList {
			if out.CanBeUnlocked(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (bc *BlockChain) GenerateSpendingPlan(address string, amount int) (int, map[string][]int) {
	// Generate a plan containing UTXOs such that the given address can use them to pay #amount to others
	// returns the total amount and plan of UTXOs
	var unspentTxs = bc.FindUnspentTransactions(address)
	var accumulated = 0
	var candidateUTXOSet = make(map[string][]int)

TxLoop:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.TxID)
		for outIdx, out := range tx.TxOutputList {
			if out.CanBeUnlocked(address) {
				accumulated += out.Value
				candidateUTXOSet[txID] = append(candidateUTXOSet[txID], outIdx)
				if accumulated >= amount {
					break TxLoop
				}
			}
		}
	}
	return accumulated, candidateUTXOSet
}

func (bc *BlockChain) GenerateTransaction(fromAddr string, toAddr string, amount int) *transaction.Transaction {

	// generate a plan of spending
	inputTotal, inputUTXOs := bc.GenerateSpendingPlan(fromAddr, amount)
	if inputTotal < amount {
		log.Panic("Error: Not enough funds!")
	}

	// create input list for new transaction
	var inputs []transaction.TxInput
	for rawTxId, OutIdxList := range inputUTXOs {
		txID, err := hex.DecodeString(rawTxId)
		utils.Handle(err)
		utils.Assert(len(OutIdxList) == 0, "Multiple TXO with same address in one transaction!")
		for _, out := range OutIdxList {
			input := transaction.TxInput{SourceTxID: txID, TxOutputIdx: out, Sig: fromAddr}
			inputs = append(inputs, input)
		}
	}

	// create output list for new transaction
	var outputs []transaction.TxOutput
	outputs = append(outputs, transaction.TxOutput{Value: amount, PubKey: toAddr})
	if inputTotal > amount {
		outputs = append(outputs, transaction.TxOutput{Value: inputTotal - amount, PubKey: fromAddr})
	}

	// create new transaction and seal it with ID
	tx := transaction.Transaction{TxInputList: inputs, TxOutputList: outputs}
	tx.SetID()
	return &tx
}

func (bc *BlockChain) Log2Terminal() {
	hasNext := true
	for iterator := bc.Iterator(); hasNext; {
		block := iterator.GetVal()
		hasNext = iterator.Next()
		block.Log2Terminal()
	}
}
