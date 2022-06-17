package blockcache

import (
	"sync"
	"github.com/AntonyMei/Blockchain/src/blocks"
	"bytes"
	"fmt"
)

type BlockCache struct {
	que []*blocks.Block
	mu sync.Mutex
	size int
	lastHash []byte
}

func InitBlockCache(size int, lastHash []byte) *BlockCache {
	c := BlockCache{size: size, lastHash: lastHash}
	// fmt.Printf("init lasthash %x.\n", c.lastHash)
	return &c
}

func (c *BlockCache) SetLastHash(lastHash []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if bytes.Compare(lastHash, c.lastHash) != 0 {
		c.lastHash = lastHash[:]
		c.que = []*blocks.Block{}
	}
	// fmt.Println("Set lasthash", c.lastHash)
}

func (c *BlockCache) AddBlock(block *blocks.Block) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	// only add when the hash is consistent
	if bytes.Compare(block.PrevHash, c.lastHash) != 0 {
		fmt.Printf("Not compatible lasthash %x %x.\n", block.PrevHash, c.lastHash)
		return false
	}
	
	// proof of work
	pow := blocks.CreateProofOfWork(block)
	if !pow.ValidateNonce() {
		fmt.Println("validate pow failed")
		return false
	}

	// check whether the block exists
	for _, cachedBlock := range c.que {
		if bytes.Compare(cachedBlock.Hash, block.Hash) != 0 {
			fmt.Println("block with same hash exists")
			return false
		}
	}

	if len(c.que) >= c.size {
		c.que = c.que[1:]
	}
	c.que = append(c.que, block)

	// fmt.Println("Added Block")

	return true
}

func (c *BlockCache) PopBlock() *blocks.Block {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.que) == 0 {
		return nil
	}
	block := c.que[0]
	c.que = c.que[1:]
	return block
}