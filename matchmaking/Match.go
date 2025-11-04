package matchmaking

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
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

func (match *Match) loop() {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()
	tick := 0
	for range ticker.C {
		tick++
		msg := struct {
			Type string `json:"type"`
			Tick int    `json:"tick"`
			Time int64  `json:"time"`
		}{
			Type: "tick",
			Tick: tick,
			Time: time.Now().UnixMilli(),
		}
		data, _ := json.Marshal(msg)
		match.broadcast(data)
	}
}
