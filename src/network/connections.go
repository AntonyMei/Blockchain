package network

import (
	"math/rand"
	"sync"
	"fmt"
)

type ConnectionPool struct {
	pool []NetworkMetaData
	mu sync.RWMutex
}

func InitializeConnectionPool() *ConnectionPool {
	cp := ConnectionPool{}
	return &cp
}

func (cp *ConnectionPool) AddPeer(peer_meta NetworkMetaData) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	for _, meta := range cp.pool {
		if(meta.Ip == peer_meta.Ip && meta.Port == peer_meta.Port) {
		  return
		}
	  }
	cp.pool = append(cp.pool, peer_meta)
}

func (cp *ConnectionPool) ExistsPeer(peer_meta NetworkMetaData) bool {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	for _, meta := range cp.pool {
		if(meta.Ip == peer_meta.Ip && meta.Port == peer_meta.Port) {
		  return true
		}
	  }
	fmt.Printf("Peer Ip=%s Port=%s doesn't exist.\n", peer_meta.Ip, peer_meta.Port)
	return false
}

func (cp *ConnectionPool) GetAlivePeers(count int) []NetworkMetaData {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	var picked_peers []NetworkMetaData
	for i := 0; i < count; i++ {
		randomIndex := rand.Intn(len(cp.pool))
		picked_peers = append(picked_peers, cp.pool[randomIndex])
	}
	return picked_peers
}

func (cp *ConnectionPool) ShowPool() {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	fmt.Printf("%d peers in connection pool\n", len(cp.pool))
	for _, peer := range cp.pool {
		fmt.Printf("    Ip = %s, Port = %s\n", peer.Ip, peer.Port)
	}
}