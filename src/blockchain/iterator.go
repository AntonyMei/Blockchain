package blockchain

import (
	"bytes"
	"github.com/AntonyMei/Blockchain/src/blocks"
	"github.com/AntonyMei/Blockchain/src/utils"
	"github.com/dgraph-io/badger"
)

type BcIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func (bc *BlockChain) Iterator() *BcIterator {
	// get last hash
	var lastHash []byte
	err := bc.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lasthash"))
		utils.Handle(err)
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		utils.Handle(err)
		return nil
	})
	utils.Handle(err)

	// create iterator
	iterator := BcIterator{lastHash, bc.Database}
	return &iterator
}

func (iterator *BcIterator) Next() bool {
	// move the iterator one step forward, return true if succeed
	// get block data
	var block *blocks.Block
	err := iterator.Database.View(func(txn *badger.Txn) error {
		// read data
		item, err := txn.Get(iterator.CurrentHash)
		utils.Handle(err)
		// reconstruct block
		err = item.Value(func(val []byte) error {
			block = blocks.Deserialize(val)
			return nil
		})
		utils.Handle(err)
		return nil
	})
	utils.Handle(err)

	// update hash pointer
	iterator.CurrentHash = block.PrevHash
	if bytes.Compare(iterator.CurrentHash, nil) == 0 {
		return false
	} else {
		return true
	}
}

func (iterator *BcIterator) GetVal() *blocks.Block {
	// get the block the iterator points to
	// get block data
	var block *blocks.Block
	err := iterator.Database.View(func(txn *badger.Txn) error {
		// read data
		item, err := txn.Get(iterator.CurrentHash)
		utils.Handle(err)
		// reconstruct block
		err = item.Value(func(val []byte) error {
			block = blocks.Deserialize(val)
			return nil
		})
		utils.Handle(err)
		return nil
	})
	utils.Handle(err)

	// return data
	return block
}
