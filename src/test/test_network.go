package test

import (
	"fmt"
	//"sync"
	"math/rand"
	"time"
	"strconv"
	"os"
	"github.com/AntonyMei/Blockchain/config"
	//"github.com/AntonyMei/Blockchain/src/blockchain"
	"github.com/AntonyMei/Blockchain/src/cli"
	"github.com/AntonyMei/Blockchain/src/network"
	//"github.com/AntonyMei/Blockchain/src/transaction"
	"github.com/AntonyMei/Blockchain/src/utils"
	//"github.com/AntonyMei/Blockchain/src/wallet"
	//"github.com/AntonyMei/Blockchain/src/cli"
)

func Init() {
    rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}

func Test_Network_Data_Bytes(num_nodes int, node_id int, Txs_Per_Block int, Outs_Per_Tx int) {

	Init()
	// construct all nodes first
	var userName string
	ip := "localhost"
	port := strconv.Itoa(5000 + node_id)
	for i := 0 ; i <= node_id; i++ {
		userName = RandStringRunes(20)
	} 

	// cli
	err := os.Mkdir(config.PersistentStoragePath+userName, os.ModePerm)
	utils.Handle(err)
	c := cli.InitializeCli(userName, ip, port)
	c.CreateWallet(userName)

	time.Sleep(time.Duration(100) * time.Millisecond)

	for {
		c.Ping("localhost", "5000")
		// broadcast all private users' id
		accountNames := c.Wallets.GetAllWalletNames()
		for _, name := range accountNames {
			wallet := c.Wallets.GetWallet(name)
			user_meta := network.UserMetaData{Name: name, PublicKey: wallet.PublicKey, WalletAddr: wallet.Address()}
			c.Node.BroadcastUserMessage(user_meta)
		}
		if len(c.Wallets.KnownAddressMap) == num_nodes {
			break
		}
		time.Sleep(time.Duration(100) * time.Millisecond)
	}

	fmt.Println("All accounts are known")

	quit := make(chan int)

	// init blocks
	go func(){
		// mine until balance is enough
		for {
			select {
			case <-quit:
				return
			default:
				c.MineBlock(userName, userName + "::" + port + "::Block" , []string{})
			}
		}
	}()
	init := false
	for {
		c.HandleBlock()

		if !init {
			// Mine untill 10000 balance for warmup
			res := c.Wallets.GetWallet(userName)
			balance := c.Blockchain.GetBalance(res.Address(), &res.PrivateKey.PublicKey)
			if balance >= 10000 {
				quit <- 1
				init = true
			}
		} else {
			c.PrintBlockchain()
			break
		}
	}


	// now i have 10000 balance

	quit = make(chan int)
	T := make(chan []string, 1)
	mined := make(chan int, 1)
	go func(){
		// mine until balance is enough
		for {
			select {
			case <-quit:
				return
			case txNameList:=<-T:
				c.MineBlock(userName, userName + "::" + port + "::Block" , txNameList)
				fmt.Printf("Mined one block\n")
				mined<-1
			}
		}
	}()

	smalltick := time.Tick(100 * time.Millisecond)
	tick := time.Tick(10 * time.Second)
	fmt.Printf("Start\n")
	mined<-1
	myTx := ""
	done := false
	for {
		select {
		case <-mined:
			if !done {
				if myTx == "" || c.PendingTxMap.GetTx(myTx) == nil{
					fmt.Printf("Generate my transaction\n")
					// create new transaction
					allKnownNames, _ := c.Wallets.GetAllKnownAddress()
					utils.Assert(len(allKnownNames) > Outs_Per_Tx, "There are not enough peers.")
					picked := make(map[string]bool)
					picked[userName] = true
					outs := []string{}
					amounts := []int{}
					for i:=0; i<Outs_Per_Tx; i++ {
						for len(outs) <= i {
							randomIndex := rand.Intn(len(allKnownNames))
							name := allKnownNames[randomIndex]
							_, exists := picked[name]
							if !exists {
								picked[name] = true
								outs = append(outs, name)
								amounts = append(amounts, 1)
							}
						}
					}
					myTx = c.CreateTransaction("tx", userName, outs, amounts)
				}
				
				fmt.Printf("Generate new transaction list\n")
				// select a new set of transactions to mine
				selectedTxs := []string{myTx}
				picked := make(map[string]bool)
				picked[myTx] = true
				allTxKeys, _ := c.PendingTxMap.GetAllTx()
				for i:=1; i<Txs_Per_Block; i++ {
					randomIndex := rand.Intn(len(allTxKeys))
					txName := allTxKeys[randomIndex]
					_, exists := picked[txName]
					if !exists {
						picked[txName] = true
						selectedTxs = append(selectedTxs, txName)
					}
				}
				T <- selectedTxs
			}
		case <-smalltick:
			// ping a random node to catchup missed block
			c.Ping("localhost", strconv.Itoa(rand.Intn(num_nodes) + 5000))
			if c.Blockchain.BlockHeight > 100 {
				// stop at 100 block height
				done = true
				for {}
			}
		case <-tick:
			// log
			c.PrintBlockchain()
			fmt.Printf("network data at block height %d: send %d bytes, receive %d bytes.\n", c.Blockchain.BlockHeight, int(c.Node.Total_send_bytes), int(c.Node.Total_recv_bytes))
		default:
			c.HandleBlock()
		}
	}
}
