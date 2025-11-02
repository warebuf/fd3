package matchmaking

import "sync"

type Match struct {
	mutex     sync.RWMutex
	ID        string
	Whitelist map[string]struct{}
	byUser    map[string]*Client
}
