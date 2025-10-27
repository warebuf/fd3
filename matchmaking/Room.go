package matchmaking

import "sync"

type Room struct {
	ID     string
	byUser map[string]*Client
	mutex  sync.RWMutex
}
