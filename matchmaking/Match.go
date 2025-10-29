package matchmaking

import "sync"

type Match struct {
	ID     string
	byUser map[string]*Client
	mutex  sync.RWMutex
}
