package match

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine/state"
)

type Match struct {
	ID        string
	Whitelist map[string]uint // acts as the teamNo map and the whitelist
	ByUser    map[string]*Client

	GameState      *state.GameState // what the game state should be 200 ms from now
	PublishedState *state.GameState // a copy of GameState that the players can read

	EventLog     []any // a log of all events that have occurred
	NextUpdateID int

	// networking related
	InputEventChan chan *InputEvent

	// EventPQ related
	EventPQueue     EventPQ
	EventPQTimer    *time.Timer
	EarliestSimItem *EventItem
	ChainItem       *EventItem
	PauseEndItem    *EventItem
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

// TODO: needs to be censored per client
func (match *Match) SendPublishedState(ClientID string) {

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

func (match *Match) BroadcastLogs(oldLogsSize int) {
	// serialize actions
	wrapper := struct {
		Type string `json:"type"`
		Data any    `json:"data"`
	}{
		Type: "actions",
		Data: match.EventLog[oldLogsSize:],
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

func (match *Match) Broadcast(data []byte) {
	for key, client := range match.ByUser {
		select {
		case client.LowPrio <- data:
		default:
			fmt.Println("dropped packet to", key)
		}
	}
}
