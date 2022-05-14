package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/AntonyMei/Blockchain/config"
	"github.com/AntonyMei/Blockchain/src/blocks"
	"github.com/AntonyMei/Blockchain/src/transaction"
	"github.com/AntonyMei/Blockchain/src/utils"
	"github.com/AntonyMei/Blockchain/src/wallet"
	"github.com/dgraph-io/badger"
	"log"
	"strconv"
)

type BlockChain struct {
	// blockchain is stored in badger database (k-v database)
	// key: hash of block, value: serialized block
	Database *badger.DB
	// Wallets: wallets of current user
	Wallets *wallet.Wallets
	// proof of difficulty
	ChainDifficulty int
	// hash of last block
	LastHash []byte
	// height of last block
	BlockHeight int
}

func InitBlockChain(wallets *wallet.Wallets, userName string) *BlockChain {
	// open db connection
	persistentPath := config.PersistentStoragePath + userName + config.BlockchainPath
	var options = badger.DefaultOptions(persistentPath)
	database, err := badger.Open(options)
	utils.Handle(err)

	// create a new blockchain if nothing exists
	blockchain := BlockChain{Database: database, Wallets: wallets, ChainDifficulty: config.InitialChainDifficulty, BlockHeight: 0}
	err = database.Update(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("lasthash"))
		if err == badger.ErrKeyNotFound {
			// no chain in database, create a new one
			fmt.Println("Initiating a new blockchain...")
			genesis := blocks.Genesis(config.InitialChainDifficulty)
			verifyResult := blockchain.ValidateBlock(genesis)
			blockchain.LastHash = genesis.Hash
			blockchain.BlockHeight = genesis.Height
			fmt.Printf("Verify genesis: %v.\n", verifyResult.String())
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

	block := blockchain.Iterator().GetVal()
	blockchain.LastHash = block.Hash
	blockchain.BlockHeight = block.Height

	return &blockchain
}

func (bc *BlockChain) AddBlock(minerAddr []byte, description string, txList []*transaction.Transaction) *blocks.Block {
	prevHeight := bc.Iterator().GetVal().Height
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

		// create new block
		txList = append(txList, transaction.CoinbaseTx(minerAddr))
		newBlock := blocks.CreateBlock(description, txList, lastHash, bc.ChainDifficulty, prevHeight)

		// check block
		verifyResult := bc.ValidateBlock(newBlock)
		fmt.Printf("Verify new block from network: %v.\n", verifyResult.String())
		if verifyResult != utils.Verified {
			return nil
		}

		// add into db
		err = txn.Set(newBlock.Hash, newBlock.Serialize())
		utils.Handle(err)
		err = txn.Set([]byte("lasthash"), newBlock.Hash)
		utils.Handle(err)
		return nil
	})
	utils.Handle(err)
	block := bc.Iterator().GetVal()
	bc.LastHash = block.Hash
	bc.BlockHeight = block.Height
	return block
}

func (bc *BlockChain) AddBlockFromNetwork(NetBlock *blocks.Block) bool {
	validBlock := false
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

		// whether this is a new block
		if bytes.Compare(NetBlock.PrevHash, lastHash) != 0 {
			// this is not not a new block
			return nil
		}

		// check block
		verifyResult := bc.ValidateBlock(NetBlock)
		fmt.Printf("Verify block: %v.\n", verifyResult.String())
		if verifyResult != utils.Verified {
			return nil
		}
		bc.LastHash = NetBlock.Hash
		bc.BlockHeight = NetBlock.Height
		validBlock = true

		// add into db
		err = txn.Set(NetBlock.Hash, NetBlock.Serialize())
		utils.Handle(err)
		err = txn.Set([]byte("lasthash"), NetBlock.Hash)
		utils.Handle(err)
		return nil
	})
	utils.Handle(err)
	return validBlock
}

func (bc *BlockChain) ValidateBlock(block *blocks.Block) utils.BlockStatus {
	wallets := bc.Wallets
	// check if this block is genesis
	if bytes.Compare(block.PrevHash, []byte{}) == 0 {
		// check hash
		pow := blocks.CreateProofOfWork(block)
		if !pow.ValidateNonce() {
			return utils.WrongGenesis
		}
		// check data
		if bytes.Compare(block.Data, []byte(config.GenesisData)) != 0 {
			return utils.WrongGenesis
		}
		// check Difficulty
		if block.Difficulty != config.InitialChainDifficulty {
			return utils.WrongGenesis
		}
		// check transactions
		if len(block.TransactionList) != 1 {
			return utils.WrongGenesis
		}
		tx := block.TransactionList[0]
		if !tx.IsCoinbase() {
			return utils.WrongGenesis
		}
		if bytes.Compare(tx.TxOutputList[0].Address, []byte(config.GenesisData)) != 0 {
			return utils.WrongGenesis
		}
		return utils.Verified
	}

	// other blocks
	// check prevHash
	var prevBlockFound bool
	err := bc.Database.View(func(txn *badger.Txn) error {
		_, err := txn.Get(block.PrevHash)
		if err == badger.ErrKeyNotFound {
			prevBlockFound = false
		} else {
			prevBlockFound = true
		}
		return nil
	})
	utils.Handle(err)
	if !prevBlockFound {
		return utils.PrevBlockNotFound
	}
	// check hash
	pow := blocks.CreateProofOfWork(block)
	if !pow.ValidateNonce() {
		return utils.HashMismatch
	}
	// get all UTXes before further checking on transactions
	var allUTXOMap = make(map[string]transaction.Transaction)
	var UTXOAddrMap = make(map[string]*wallet.KnownAddress)
	for _, knownAddr := range wallets.KnownAddressMap {
		tmpUTXList := bc.FindUnspentTransactions(knownAddr.Address, &knownAddr.PublicKey)
		for _, tmpUTX := range tmpUTXList {
			for outIdx, out := range tmpUTX.TxOutputList {
				if out.BelongsTo(knownAddr.Address) {
					allUTXOMap[string(tmpUTX.TxID)+strconv.Itoa(outIdx)] = tmpUTX
					UTXOAddrMap[string(tmpUTX.TxID)+strconv.Itoa(outIdx)] = knownAddr
				}
			}
		}
	}
	// check transactions
	coinbaseTXCount := 0
	var SpentUXTOMap = make(map[string]bool)
	for _, tx := range block.TransactionList {
		// check if it is coinbase TX
		if tx.IsCoinbase() {
			coinbaseTXCount += 1
			if coinbaseTXCount > 1 {
				return utils.TooManyCoinbaseTX
			}
			continue
		}
		// check if TxID is correct
		txCopy := transaction.Transaction{TxInputList: tx.TxInputList, TxOutputList: tx.TxOutputList}
		txCopy.SetID()
		if bytes.Compare(txCopy.TxID, tx.TxID) != 0 {
			return utils.WrongTxID
		}
		// check each input of TX
		inputSum := 0
		for _, txInput := range tx.TxInputList {
			// check whether the source TXO exists in UTXO set
			sourceTx, exists := allUTXOMap[string(txInput.SourceTxID)+strconv.Itoa(txInput.TxOutputIdx)]
			if !exists {
				return utils.SourceTXONotFound
			}
			// check whether the input is correctly signed
			signer := UTXOAddrMap[string(txInput.SourceTxID)+strconv.Itoa(txInput.TxOutputIdx)]
			if !txInput.Verify(&signer.PublicKey) {
				return utils.WrongTXInputSignature
			}
			_, exists = SpentUXTOMap[string(txInput.SourceTxID)+strconv.Itoa(txInput.TxOutputIdx)]
			if !exists {
				SpentUXTOMap[string(txInput.SourceTxID)+strconv.Itoa(txInput.TxOutputIdx)] = true
			} else {
				return utils.DoubleSpending
			}
			// accumulate to inputSum
			inputSum += sourceTx.TxOutputList[txInput.TxOutputIdx].Value
		}
		// check if sum of input is equal to sum of output
		outputSum := 0
		for _, txOutput := range tx.TxOutputList {
			outputSum += txOutput.Value
		}
		if outputSum != inputSum {
			return utils.InputSumOutputSumMismatch
		}
	}
	return utils.Verified
}

func (bc *BlockChain) FindUnspentTransactions(address []byte, publicKey *ecdsa.PublicKey) []transaction.Transaction {
	// This function returns all transactions that contain unspent outputs associated with a wallet

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
				if out.BelongsTo(address) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			// mark all its inputs as spent
			if tx.IsCoinbase() == false {
				for _, in := range tx.TxInputList {
					if in.Verify(publicKey) {
						inTxID := hex.EncodeToString(in.SourceTxID)
						spentTxMap[inTxID] = append(spentTxMap[inTxID], in.TxOutputIdx)
					}
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return unspentTxs
}

func (bc *BlockChain) FindUTXO(address []byte, publicKey *ecdsa.PublicKey) []transaction.TxOutput {
	// This function returns all UTXOs associated with address
	var UTXOs []transaction.TxOutput
	unspentTransactions := bc.FindUnspentTransactions(address, publicKey)
	for _, tx := range unspentTransactions {
		for _, out := range tx.TxOutputList {
			if out.BelongsTo(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (bc *BlockChain) GenerateSpendingPlan(wallet *wallet.Wallet, amount int) (int, map[string][]int) {
	// Generate a plan containing UTXOs such that the given address can use them to pay #amount to others
	// returns the total amount and plan of UTXOs
	var unspentTxs = bc.FindUnspentTransactions(wallet.Address(), &wallet.PrivateKey.PublicKey)
	var accumulated = 0
	var candidateUTXOSet = make(map[string][]int)
	var address = wallet.Address()

TxLoop:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.TxID)
		for outIdx, out := range tx.TxOutputList {
			if out.BelongsTo(address) {
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

func (bc *BlockChain) GenerateTransaction(fromWallet *wallet.Wallet, toAddrList [][]byte, amountList []int) *transaction.Transaction {
	// generate a transaction
	// check input
	utils.Assert(len(toAddrList) == len(amountList), "TX error: receiver and amount dimension mismatch.")

	// generate a plan of spending
	totalAmount := 0
	for _, amount := range amountList {
		totalAmount += amount
	}
	inputTotal, inputUTXOs := bc.GenerateSpendingPlan(fromWallet, totalAmount)
	if inputTotal < totalAmount {
		log.Panic("Error: Not enough funds!")
	}

	// create input list for new transaction
	var inputs []transaction.TxInput
	for rawTxId, OutIdxList := range inputUTXOs {
		txID, err := hex.DecodeString(rawTxId)
		utils.Handle(err)
		utils.Assert(len(OutIdxList) == 1, "Multiple TXO with same address in one transaction!")
		for _, out := range OutIdxList {
			input := transaction.TxInput{SourceTxID: txID, TxOutputIdx: out}
			input.Sign(&fromWallet.PrivateKey)
			inputs = append(inputs, input)
		}
	}

	// create output list for new transaction
	var outputs []transaction.TxOutput
	for idx := range toAddrList {
		outputs = append(outputs, transaction.TxOutput{Value: amountList[idx], Address: toAddrList[idx]})
	}
	if inputTotal > totalAmount {
		outputs = append(outputs, transaction.TxOutput{Value: inputTotal - totalAmount, Address: fromWallet.Address()})
	}

	// create new transaction and seal it with ID
	tx := transaction.Transaction{TxInputList: inputs, TxOutputList: outputs}
	tx.SetID()
	return &tx
}

func (bc *BlockChain) GetBalance(address []byte, publicKey *ecdsa.PublicKey) int {
	// Get balance of an account
	var unspentTxs = bc.FindUnspentTransactions(address, publicKey)
	var balance = 0
	for _, tx := range unspentTxs {
		for _, out := range tx.TxOutputList {
			if out.BelongsTo(address) {
				balance += out.Value
			}
		}
	}
	return balance
}

func (bc *BlockChain) Log2Terminal() {
	hasNext := true
	for iterator := bc.Iterator(); hasNext; {
		block := iterator.GetVal()
		hasNext = iterator.Next()
		block.Log2Terminal()
	}
}

func (bc *BlockChain) Exit() {
	err := bc.Database.Close()
	utils.Handle(err)
}
