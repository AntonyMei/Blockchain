package pow

import (
	"bytes"
	"crypto/sha256"
	"github.com/AntonyMei/Blockchain/src/basic"
	"github.com/AntonyMei/Blockchain/src/utils"
	"math"
	"math/big"
)

type ProofOfWork struct {
	Block  *basic.Block
	Target *big.Int
}

func CreateProofOfWork(block *basic.Block) *ProofOfWork {
	// we use target to ensure that the high bits of hash are 0
	target := big.NewInt(1)
	target.Lsh(target, uint(256-block.Difficulty))
	pow := &ProofOfWork{block, target}
	return pow
}

func (pow *ProofOfWork) GenerateNonce() (int, []byte) {
	// return nonce, hash
	var intHash big.Int
	var hash [32]byte
	for nonce := 0; nonce < math.MaxInt64; {
		powData := bytes.Join([][]byte{pow.Block.PrevHash, pow.Block.Data,
			utils.Int2Hex(int64(nonce)),
			utils.Int2Hex(int64(pow.Block.Difficulty))}, []byte{})
		hash = sha256.Sum256(powData)
		intHash.SetBytes(hash[:])
		if intHash.Cmp(pow.Target) == -1 {
			return nonce, hash[:]
		} else {
			nonce++
		}
	}
	panic("Nonce not found")
}

func (pow *ProofOfWork) ValidateNonce() bool {
	// check that nonce can really make initial bits of hash value 0
	var intHash big.Int
	powData := bytes.Join([][]byte{pow.Block.PrevHash, pow.Block.Data,
		utils.Int2Hex(int64(pow.Block.Nonce)),
		utils.Int2Hex(int64(pow.Block.Difficulty))}, []byte{})
	hash := sha256.Sum256(powData)
	intHash.SetBytes(hash[:])
	return intHash.Cmp(pow.Target) == -1
}
