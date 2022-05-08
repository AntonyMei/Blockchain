package blockchain

import (
	"fmt"
	"github.com/AntonyMei/Blockchain/src/blocks"
	"github.com/AntonyMei/Blockchain/src/utils"
	"github.com/dgraph-io/badger"
)

// InitialChainDifficulty = #zeros at hash head * 4
const InitialChainDifficulty = 16

// PersistentStoragePath is where we store the chain on disk
const PersistentStoragePath = "./tmp/blocks"

type BlockChain struct {
	// blockchain is stored in badger database (k-v database)
	// key: hash of block, value: serialized block
	Database *badger.DB
	// proof of difficulty
	ChainDifficulty int
}

func InitBlockChain() *BlockChain {
	// open db connection
	var options = badger.DefaultOptions(PersistentStoragePath)
	database, err := badger.Open(options)
	utils.Handle(err)

	// create a new blockchain if nothing exists
	_ = database.Update(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("lasthash"))
		if err == badger.ErrKeyNotFound {
			// no chain in database, create a new one
			fmt.Println("Initiating a new blockchain...")
			genesis := blocks.Genesis(InitialChainDifficulty)
			err = txn.Set(genesis.Hash, genesis.Serialize())
			utils.Handle(err)
			err = txn.Set([]byte("lasthash"), genesis.Hash)
			utils.Handle(err)
		}
		return nil
	})

	blockchain := BlockChain{Database: database, ChainDifficulty: InitialChainDifficulty}
	return &blockchain
}

func (bc *BlockChain) AddBlock(data string) {
	//newBlock := blocks.CreateBlock(data, bc.BlockList[len(bc.BlockList)-1].Hash, bc.ChainDifficulty)
	//bc.BlockList = append(bc.BlockList, newBlock)
}
