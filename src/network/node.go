package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net/http"
    "net/url"
	"time"
	"sync"
	"github.com/AntonyMei/Blockchain/src/utils"
	"github.com/AntonyMei/Blockchain/src/wallet"
	"github.com/AntonyMei/Blockchain/src/blocks"
	"github.com/AntonyMei/Blockchain/src/blockchain"
	"github.com/AntonyMei/Blockchain/src/transaction"
)

type Node struct {
	ConnectionPool *ConnectionPool
	Wallets *wallet.Wallets
	Chain *blockchain.BlockChain
	Meta NetworkMetaData
	mu sync.Mutex
	Blocks []*blocks.Block
	CliHandleTxFromNetwork func(string, *transaction.Transaction)
	CliHandleBlockFromNetwork func(*blocks.Block)
}

func InitializeNode(w *wallet.Wallets, chain *blockchain.BlockChain, meta NetworkMetaData) *Node {
	nd := Node{ConnectionPool: InitializeConnectionPool(), Wallets: w, Chain: chain, Meta: meta}
	nd.ConnectionPool.AddPeer(nd.Meta)
	return &nd
}

func (nd *Node) SetCliTransactionFunc(f func(string, *transaction.Transaction)) {
	nd.CliHandleTxFromNetwork = f
}

func (nd *Node) SetCliBlockFunc(f func(*blocks.Block)) {
	nd.CliHandleBlockFromNetwork = f
}

func PrintHeader(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
        for _, h := range headers {
            fmt.Fprintf(w, "%v: %v\n", name, h)
        }
    }
}

func (nd *Node) AddBlock(newBlock *blocks.Block) {
	if newBlock == nil {
		return
	}
	nd.mu.Lock()
	defer nd.mu.Unlock()
	// if already exists, do not add the block
	for _, block := range nd.Blocks {
		if bytes.Compare(block.Hash, newBlock.Hash) == 0 {
			return
		}
	}
	nd.Blocks = append(nd.Blocks, newBlock)
}

func (nd *Node) GetBlock(blockHeight int) *blocks.Block {
	nd.mu.Lock()
	defer nd.mu.Unlock()
	for _, block := range nd.Blocks {
		if block.Height == blockHeight {
			return block
		}
	}
	return nil
}

func (nd *Node) HandlePingMessage(w http.ResponseWriter, req *http.Request) {
	PrintHeader(w, req)

	// send acknowledgement back
	fmt.Fprintf(w, "ACK")

	body, err := ioutil.ReadAll(req.Body)
	utils.Handle(err)

	var msg PingMessage
	var decoder = gob.NewDecoder(bytes.NewReader(body))
	utils.Handle(decoder.Decode(&msg))

	// fmt.Printf("Receive PING message from http://%s:%s with block height %d.\n", msg.Meta.Ip, msg.Meta.Port, msg.BlockHeight)
	
	nd.SendPeersMessage(msg.Meta)
	
	if nd.ConnectionPool.AddPeer(msg.Meta) {
		nd.SendPingMessage(msg.Meta, nd.Chain.BlockHeight)
	}

	// synchronize block according to block height
	if msg.BlockHeight < nd.Chain.BlockHeight {
		block := nd.GetBlock(msg.BlockHeight + 1)
		if block != nil {
			nd.SendBlockMessage(msg.Meta, block)
		}
	}
}

func (nd *Node) HandlePeersMessage(w http.ResponseWriter, req *http.Request) {
	PrintHeader(w, req)

	// send acknowledgement back
	fmt.Fprintf(w, "ACK")
	body, err := ioutil.ReadAll(req.Body)
	utils.Handle(err)

	var msg PeersMessage
	var decoder = gob.NewDecoder(bytes.NewReader(body))
	utils.Handle(decoder.Decode(&msg))

	// fmt.Printf("Receive PEERS message from http://%s:%s.\n", msg.Meta.Ip, msg.Meta.Port)

	for _, peer := range msg.Peers {
		if nd.ConnectionPool.AddPeer(peer) {
			nd.SendPingMessage(peer, nd.Chain.BlockHeight)
		}
	}
}

func (nd *Node) HandleUserMessage(w http.ResponseWriter, req *http.Request) {
	PrintHeader(w, req)

	// send acknowledgement back
	fmt.Fprintf(w, "ACK")
	body, err := ioutil.ReadAll(req.Body)
	utils.Handle(err)

	var msg UserMessage
	var decoder = gob.NewDecoder(bytes.NewReader(body))
	utils.Handle(decoder.Decode(&msg))

	// fmt.Printf("Receive USER message from http://%s:%s. Name=%s\n", msg.Meta.Ip, msg.Meta.Port, msg.UserMeta.Name)

	nd.Wallets.AddKnownAddress(msg.UserMeta.Name, &wallet.KnownAddress{PublicKey: wallet.DeserializePublicKey(msg.UserMeta.PublicKey), Address: msg.UserMeta.WalletAddr})
}

func (nd *Node) HandleTransactionMessage(w http.ResponseWriter, req *http.Request) {
	PrintHeader(w, req)

	// send acknowledgement back
	fmt.Fprintf(w, "ACK")
	body, err := ioutil.ReadAll(req.Body)
	utils.Handle(err)

	var msg TransactionMessage
	var decoder = gob.NewDecoder(bytes.NewReader(body))
	utils.Handle(decoder.Decode(&msg))
	
	//fmt.Printf("Get Transaction from Ip=%s Port=%s.\n", msg.Meta.Ip, msg.Meta.Port)

	txKey := msg.TxKey
	tx := msg.Transaction
	nd.CliHandleTxFromNetwork(txKey, tx)
}

func (nd *Node) HandleBlockMessage(w http.ResponseWriter, req *http.Request) {
	PrintHeader(w, req)

	// send acknowledgement back
	fmt.Fprintf(w, "ACK")
	body, err := ioutil.ReadAll(req.Body)
	utils.Handle(err)

	var msg BlockMessage
	var decoder = gob.NewDecoder(bytes.NewReader(body))
	utils.Handle(decoder.Decode(&msg))
	
	//fmt.Printf("Get Block from Ip=%s Port=%s.\n", msg.Meta.Ip, msg.Meta.Port)

	nd.CliHandleBlockFromNetwork(msg.Block)
}

func (nd *Node) SendMessage(channel string, meta NetworkMetaData, buf *bytes.Buffer) {
	c := http.Client{Timeout: time.Duration(10) * time.Second}

	s := fmt.Sprintf("http://%s:%s/%s", meta.Ip, meta.Port, channel)
	url, err := url.Parse(s)
	utils.Handle(err)

	resp, err := c.Post(url.String(), "", bytes.NewBuffer(buf.Bytes()))
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	utils.Assert((string(body)[len(string(body))-3:] == "ACK"), fmt.Sprintf("response = ACK, but get %s\n", body))
	defer resp.Body.Close()
}

func (nd *Node) SendPeersMessage(meta NetworkMetaData) {
	peers := nd.ConnectionPool.GetAlivePeers(20)

	msg := CreatePeersMessage(nd.Meta, peers)
	var result bytes.Buffer
	var encoder = gob.NewEncoder(&result)
	utils.Handle(encoder.Encode(msg))

	nd.SendMessage("peers", meta, &result)
}

func (nd *Node) SendPingMessage(meta NetworkMetaData, blockHeight int) {
	msg := CreatePingMessage(nd.Meta, blockHeight)
	var result bytes.Buffer
	var encoder = gob.NewEncoder(&result)
	utils.Handle(encoder.Encode(msg))

	nd.SendMessage("ping", meta, &result)
}

func (nd *Node) SendBlockMessage(meta NetworkMetaData, block *blocks.Block) {
	msg := CreateBlockMessage(nd.Meta, block)
	var result bytes.Buffer
	var encoder = gob.NewEncoder(&result)
	utils.Handle(encoder.Encode(msg))

	nd.SendMessage("block", meta, &result)
}

func (nd *Node) BroadcastUserMessage(userMeta UserMetaData) {
	peers := nd.ConnectionPool.GetAlivePeers(50)
	
	msg := CreateUserMessage(nd.Meta, userMeta)
	var result bytes.Buffer
	var encoder = gob.NewEncoder(&result)
	utils.Handle(encoder.Encode(msg))

	var SentPeer = make(map[NetworkMetaData]bool)
	for _, peer := range peers {
		_, exist := SentPeer[peer]
		if !exist {
			SentPeer[peer] = true
			nd.SendMessage("user", peer, &result)
		}
	}
}

func (nd *Node) BroadcastTransaction(txKey string, tx *transaction.Transaction) {
	//fmt.Printf("Broadcast transaction %s.\n", txKey)
	peers := nd.ConnectionPool.GetAlivePeers(50)

	msg := CreateTransactionMessage(nd.Meta, txKey, tx)
	var result bytes.Buffer
	var encoder = gob.NewEncoder(&result)
	utils.Handle(encoder.Encode(msg))

	var SentPeer = make(map[NetworkMetaData]bool)
	for _, peer := range peers {
		_, exist := SentPeer[peer]
		if !exist {
			SentPeer[peer] = true
			nd.SendMessage("transaction", peer, &result)
		}
	}

}

func (nd *Node) BroadcastBlock(block *blocks.Block) {
	if block == nil {
		return
	}
	//fmt.Printf("Broadcast block with Hash %x.\n", block.Hash)
	peers := nd.ConnectionPool.GetAlivePeers(50)
	
	msg := CreateBlockMessage(nd.Meta, block)
	var result bytes.Buffer
	var encoder = gob.NewEncoder(&result)
	utils.Handle(encoder.Encode(msg))

	var SentPeer = make(map[NetworkMetaData]bool)
	for _, peer := range peers {
		_, exist := SentPeer[peer]
		if !exist {
			//fmt.Printf("Send block to Ip=%s, Port=%s.\n", peer.Ip, peer.Port)
			SentPeer[peer] = true
			nd.SendMessage("block", peer, &result)
		}
	}
}

func (nd *Node) Serve() error {
	http.HandleFunc("/ping", nd.HandlePingMessage)
	http.HandleFunc("/peers", nd.HandlePeersMessage)
	http.HandleFunc("/user", nd.HandleUserMessage)
	http.HandleFunc("/block", nd.HandleBlockMessage)
	http.HandleFunc("/transaction", nd.HandleTransactionMessage)

	go func() {
		fmt.Printf("Listening at port %s\n", nd.Meta.Port)
		err := http.ListenAndServe(fmt.Sprintf(":%s", nd.Meta.Port), nil)
		utils.Handle(err)
	}()
	return nil
}