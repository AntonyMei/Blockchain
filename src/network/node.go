package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net/http"
    "net/url"
	"time"
	"github.com/AntonyMei/Blockchain/src/utils"
	"github.com/AntonyMei/Blockchain/src/wallet"
	"github.com/AntonyMei/Blockchain/src/blockchain"
)

type Node struct {
	ConnectionPool *ConnectionPool
	Wallets *wallet.Wallets
	Chain *blockchain.BlockChain
	Meta NetworkMetaData
}

func InitializeNode(w *wallet.Wallets, chain *blockchain.BlockChain, meta NetworkMetaData) *Node {
	nd := Node{InitializeConnectionPool(), w, chain, meta}
	return &nd
}

func PrintHeader(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
        for _, h := range headers {
            fmt.Fprintf(w, "%v: %v\n", name, h)
        }
    }
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

	fmt.Printf("Receive PING message from http://%s:%s.\n", msg.Meta.Ip, msg.Meta.Port)
	
	nd.ConnectionPool.AddPeer(msg.Meta)
	nd.SendPeersMessage(msg.Meta)

	// TODO: synchronize block according to block height
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

	nd.ConnectionPool.AddPeer(msg.Meta)
	fmt.Printf("Receive PEERS message from http://%s:%s.\n", msg.Meta.Ip, msg.Meta.Port)

	for _, peer := range msg.Peers {
		if !nd.ConnectionPool.ExistsPeer(peer) {
			nd.ConnectionPool.AddPeer(peer)
			nd.SendPingMessage(peer)
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

	fmt.Printf("Receive USER message from http://%s:%s. Name=%s\n", msg.Meta.Ip, msg.Meta.Port, msg.UserMeta.Name)

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
	
	/*TODO: add transaction into block chain

	tx := msg.Transaction

	nd.Chain.AddTransaction(tx)*/
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
	
	/*TODO: add block into block chain
	1. validate block & nounce
	2. validate transaction in comming block
	3. Recompute unspent transactions & revalidate transactions
	*/
}

func (nd *Node) SendMessage(channel string, meta NetworkMetaData, buf *bytes.Buffer) {
	c := http.Client{Timeout: time.Duration(10) * time.Second}

	s := fmt.Sprintf("http://%s:%s/%s", meta.Ip, meta.Port, channel)
	url, err := url.Parse(s)
	utils.Handle(err)

	resp, err := c.Post(url.String(), "", bytes.NewBuffer(buf.Bytes()))
	utils.Handle(err)
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

func (nd *Node) SendPingMessage(meta NetworkMetaData) {

	// TODO support BlockHeight
	msg := CreatePingMessage(nd.Meta, -1)
	var result bytes.Buffer
	var encoder = gob.NewEncoder(&result)
	utils.Handle(encoder.Encode(msg))

	nd.SendMessage("ping", meta, &result)
}

func (nd *Node) BroadcastUserMessage(userMeta UserMetaData) {
	peers := nd.ConnectionPool.GetAlivePeers(20)
	
	msg := CreateUserMessage(nd.Meta, userMeta)
	var result bytes.Buffer
	var encoder = gob.NewEncoder(&result)
	utils.Handle(encoder.Encode(msg))

	for _, peer := range peers {
		nd.SendMessage("user", peer, &result)
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