package blockchain

import (
	"fmt"
	"sync"
	"github.com/AntonyMei/Blockchain/src/transaction"
)

type PendingTXs struct {
	pendingTXMap map[string]*transaction.Transaction
	mu sync.Mutex
}

func InitPendingTXs() *PendingTXs {
	var p PendingTXs
	p.pendingTXMap = make(map[string]*transaction.Transaction)
	return &p
}

func (p *PendingTXs) AddTransaction(txKey string, tx *transaction.Transaction) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pendingTXMap[txKey] = tx
}

func (p *PendingTXs) ListPendingTransactions() { 
	p.mu.Lock()
	defer p.mu.Unlock()
	idx := 0
	for txKey, tx := range p.pendingTXMap {
		fmt.Printf("Transaction %v: %s\n", idx, txKey)
		tx.Log2Terminal()
		idx += 1
	}
}

func (p *PendingTXs) GetTx(txKey string) *transaction.Transaction {
	p.mu.Lock()
	tx := p.pendingTXMap[txKey]
	p.mu.Unlock()
	return tx
}

func (p *PendingTXs) DeleteTx(txKey string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.pendingTXMap, txKey)
}

func (p *PendingTXs) GetAllTx() ([]string, []*transaction.Transaction) {
	p.mu.Lock()
	allTxs := []*transaction.Transaction{}
	allTxKeys := []string{}
	for txKey, tx := range p.pendingTXMap {
		allTxs = append(allTxs, tx)
		allTxKeys = append(allTxKeys, txKey)
	}
	p.mu.Unlock()
	return allTxKeys, allTxs
}
