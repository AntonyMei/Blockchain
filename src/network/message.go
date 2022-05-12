package network

import (
	"github.com/AntonyMei/Blockchain/src/transaction"
	"github.com/AntonyMei/Blockchain/src/blocks"
)

type NetworkMetaData struct {
	Ip string
	Port string
	Name string
	PublicKey []byte
	WalletAddr []byte
}

type PingMessage struct {
	Meta NetworkMetaData
	BlockHeight int
}

func CreatePingMessage(Meta NetworkMetaData, BlockHeight int) PingMessage {
	msg := PingMessage{Meta, BlockHeight}
	return msg
} 

type PeersMessage struct {
	Meta NetworkMetaData
	Peers []NetworkMetaData
}

func CreatePeersMessage(Meta NetworkMetaData, Peers []NetworkMetaData) PeersMessage {
	msg := PeersMessage{Meta, Peers}
	return msg
}

type TransactionMessage struct {
	Meta NetworkMetaData
	Transaction transaction.Transaction
}

func CreateTransactionMessage(Meta NetworkMetaData, Transaction transaction.Transaction) TransactionMessage {
	msg := TransactionMessage{Meta, Transaction}
	return msg
}

type BlockMessage struct {
	Meta NetworkMetaData
	Block *blocks.Block
}

func CreateBlockMessage(Meta NetworkMetaData, Block *blocks.Block) BlockMessage {
	msg := BlockMessage{Meta, Block}
	return msg
}