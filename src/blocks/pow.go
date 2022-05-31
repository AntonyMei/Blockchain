package blocks

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/AntonyMei/Blockchain/src/utils"
	"math"
	"math/big"
	"time"
)

type ProofOfWorkWrapper struct {
	Block  *Block
	Target *big.Int
}

func CreateProofOfWork(block *Block) *ProofOfWorkWrapper {
	// we use target to ensure that the high bits of hash are 0
	target := big.NewInt(1)
	target.Lsh(target, uint(256-block.Difficulty))
	pow := &ProofOfWorkWrapper{block, target}
	return pow
}

func FindNonce(pow *ProofOfWorkWrapper, workId int, totalWorker int,
	resultChan chan int, killSigChan chan struct{}, workloadChan chan int) {
	var intHash big.Int
	var hash [32]byte
	chunkSize := int64(math.MaxInt64 / totalWorker)
	for nonce := int64(workId) * chunkSize; nonce < int64(workId+1)*chunkSize; {
		select {
		case <-killSigChan:
			println("killed")
			workloadChan <- int(nonce)
			return
		default:
			powData := bytes.Join([][]byte{pow.Block.PrevHash, pow.Block.Data,
				pow.Block.GetTransactionsHash(),
				utils.Int2Hex(int64(nonce)),
				utils.Int2Hex(int64(pow.Block.Difficulty))}, []byte{})
			hash = sha256.Sum256(powData)
			intHash.SetBytes(hash[:])
			if intHash.Cmp(pow.Target) == -1 {
				resultChan <- int(nonce)
				workloadChan <- int(nonce)
				println("found!")
				return
			} else {
				nonce++
			}
		}
	}
}

func (pow *ProofOfWorkWrapper) GenerateNonceHash() (int, []byte) {
	// Spawn goroutines to find nonce
	start := time.Now().UnixMilli()
	//cpuNum := runtime.NumCPU()
	cpuNum := 4
	routineNum := int(math.Max(1, float64(cpuNum-4)))
	resultChan := make(chan int)
	workloadChan := make(chan int)
	killSigChan := make(chan struct{})
	defer close(resultChan)
	for i := 0; i < routineNum; i++ {
		go FindNonce(pow, i, routineNum, resultChan, killSigChan, workloadChan)
	}
	nonce := <-resultChan
	close(killSigChan) // This will kill all go routines
	end := time.Now().UnixMilli()

	// calculate total work
	totalWorkload := 0
WorkloadLoop:
	for {
		select {
		case workload := <-workloadChan:
			totalWorkload += workload
		default:
			break WorkloadLoop
		}
	}

	// return nonce, hash
	var intHash big.Int
	var hash [32]byte
	powData := bytes.Join([][]byte{pow.Block.PrevHash, pow.Block.Data,
		pow.Block.GetTransactionsHash(),
		utils.Int2Hex(int64(nonce)),
		utils.Int2Hex(int64(pow.Block.Difficulty))}, []byte{})
	hash = sha256.Sum256(powData)
	intHash.SetBytes(hash[:])
	if intHash.Cmp(pow.Target) == -1 {
		// a million hash per second
		hashRate := (float64(totalWorkload) / float64(end-start)) / 1000
		fmt.Printf("Hash rate: %fMH/s.\n", hashRate)
		return nonce, hash[:]
	} else {
		panic("Wrong nonce returned by worker!")
	}
}

func (pow *ProofOfWorkWrapper) ValidateNonce() bool {
	// check that nonce can really make initial bits of hash value 0
	var intHash big.Int
	powData := bytes.Join([][]byte{pow.Block.PrevHash, pow.Block.Data,
		pow.Block.GetTransactionsHash(),
		utils.Int2Hex(int64(pow.Block.Nonce)),
		utils.Int2Hex(int64(pow.Block.Difficulty))}, []byte{})
	hash := sha256.Sum256(powData)
	intHash.SetBytes(hash[:])
	return intHash.Cmp(pow.Target) == -1
}
