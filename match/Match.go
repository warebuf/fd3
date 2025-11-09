package match

import (
	"fmt"
	"sync"
)

type Match struct {
	mutex     sync.RWMutex
	ID        string
	Whitelist map[string]struct{}
	ByUser    map[string]*Client
}

func (match *Match) broadcast(data []byte) {
	for key, client := range match.ByUser {
		select {
		case client.LowPrio <- data:
		default:
			fmt.Println("dropped packet to", key)
		}
	}
}
