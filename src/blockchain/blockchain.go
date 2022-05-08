package blockchain

import (
	"fmt"
	"github.com/AntonyMei/Blockchain/config"
	"github.com/AntonyMei/Blockchain/src/blocks"
	"github.com/AntonyMei/Blockchain/src/transaction"
	"github.com/AntonyMei/Blockchain/src/utils"
	"github.com/dgraph-io/badger"
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
			coinbaseTx := transaction.CoinbaseTx(genesisMinerAddr, config.GenesisTxData)
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

func (bc *BlockChain) AddBlock(data string, txList []*transaction.Transaction) {
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
		newBlock := blocks.CreateBlock(data, txList, lastHash, bc.ChainDifficulty)
		err = txn.Set(newBlock.Hash, newBlock.Serialize())
		utils.Handle(err)
		err = txn.Set([]byte("lasthash"), newBlock.Hash)
		utils.Handle(err)
		return nil
	})
	utils.Handle(err)
}
