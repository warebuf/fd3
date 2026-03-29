package match

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/simulate"
)

type Match struct {
	ID        string
	Whitelist map[string]uint // acts as the teamNo map and the whitelist
	ByUser    map[string]*Client

	GameState      *engine.GameState       // what the game state should be 200 ms from now
	PublishedState *engine.GameState       // a copy of GameState that the players can read
	PQ             *simulate.PriorityQueue // priority queue of chains/stacks
	NextNodeIndex  uint64                  // assign an index to each action/effect
	EventLog       []engine.EventLog       // a log of all events that have occurred

	NextRoundTime time.Time

	InputEventChan chan *InputEvent
	EventPQueue    EventPQ
	EventPQTimer   *time.Timer
	NextSeq        int
}

// TODO: needs to be censored per client
func (match *Match) BroadcastPublishedState() {

	// serialize game state
	wrapper := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: "fullState",
		Data: match.PublishedState,
	}
	gameStateJSON, err := json.Marshal(wrapper)
	if err != nil {
		fmt.Println("problem accessing game state", err)
		return
	}

	// send to all clients
	for key, client := range match.ByUser {
		select {
		case client.LowPrio <- gameStateJSON:
		default:
			fmt.Println("dropped packet to", key)
		}
	}
}

func (match *Match) BroadcastEventLog(acts []engine.EventLog) {
	// serialize actions
	wrapper := struct {
		Type string `json:"type"`
		Data any    `json:"data"`
	}{
		Type: "actions",
		Data: acts,
	}

	payload, err := json.Marshal(wrapper)
	if err != nil {
		fmt.Println("problem marshaling actions:", err)
		return
	}

	// send to all clients
	for key, client := range match.ByUser {
		select {
		case client.LowPrio <- payload:
		default:
			fmt.Println("dropped packet to", key)
		}
	}
}

// TODO: needs to be censored per client
func (match *Match) SendGameState(ClientID string) {

	// serialize current game state
	wrapper := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: "fullState",
		Data: match.PublishedState,
	}
	gameStateJSON, err := json.Marshal(wrapper)
	if err != nil {
		fmt.Println("problem accessing game state", err)
		return
	}

	// send it to the client
	if clientObj, ok := match.ByUser[ClientID]; ok {
		select {
		case clientObj.LowPrio <- gameStateJSON:
		default:
			fmt.Println("SendGameState, dropped packet to", ClientID)
		}
	}

}

func (match *Match) Broadcast(data []byte) {
	for key, client := range match.ByUser {
		select {
		case client.LowPrio <- data:
		default:
			fmt.Println("dropped packet to", key)
		}
	}
}
