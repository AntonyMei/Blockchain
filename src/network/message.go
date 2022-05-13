package network

import (
	"github.com/AntonyMei/Blockchain/src/transaction"
	"github.com/AntonyMei/Blockchain/src/blocks"
)

type NetworkMetaData struct {
	Ip string
	Port string
}

type UserMetaData struct {
	Name string
	PublicKey []byte
	WalletAddr []byte
}

type UserMessage struct {
	Meta NetworkMetaData
	UserMeta UserMetaData
}

func CreateUserMessage(Meta NetworkMetaData, UserMeta UserMetaData) UserMessage {
	msg := UserMessage{Meta, UserMeta}
	return msg
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
	TxKey string
	Transaction *transaction.Transaction
}

func CreateTransactionMessage(Meta NetworkMetaData, txKey string, Transaction *transaction.Transaction) TransactionMessage {
	msg := TransactionMessage{Meta, txKey, Transaction}
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