package network

import (
	"math/rand"
	"sync"
)

type ConnectionPool struct {
	pool map[string]NetworkMetaData
	mu sync.RWMutex
}

func InitializeConnectionPool() *ConnectionPool {
	cp := ConnectionPool{}
	cp.pool = make(map[string]NetworkMetaData)
	return &cp
}

func (cp *ConnectionPool) AddPeer(peer_meta NetworkMetaData) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.pool[peer_meta.Name] = peer_meta
}

func (cp *ConnectionPool) ExistsPeer(peer_meta NetworkMetaData) bool {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	for name, meta := range cp.pool {
		if(name == peer_meta.Name && meta.Ip == peer_meta.Ip && meta.Port == peer_meta.Port) {
		  return true
		}
	  }
	return false
}

func (cp *ConnectionPool) GetAlivePeers(count int) []NetworkMetaData {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	var all_peers []string
	for name, _ := range cp.pool {
		all_peers = append(all_peers, name)
	  }
	
	var picked_peers []NetworkMetaData
	for i := 0; i < count; i++ {
		randomIndex := rand.Intn(len(all_peers))
    	peer_name := all_peers[randomIndex]
		picked_peers = append(picked_peers, cp.pool[peer_name])
	}
	return picked_peers
}