package match

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/warebuf/fd3/debug"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

type Match struct {
	ID        string
	Whitelist map[string]uint // acts as the teamNo map and the whitelist
	ByUser    map[string]*Client

	GameState      *state.GameState // what the game state should be 200 ms from now
	PublishedState *state.GameState // TODO: replace with copy upon request

	SimulationPQ state.SimulationPQ

	EventLog     []any // a log of all events that have occurred
	NextUpdateID int
	DebugTrace   []debug.TraceEntry

	// networking related
	InputEventChan chan *eventpq.InputEvent

	// EventPQ related
	EventPQueue  eventpq.EventPQ
	EventPQTimer *time.Timer

	EarliestSimItem *eventpq.EventItem
	LatestChainItem *eventpq.EventItem
	PauseEndItem    *eventpq.EventItem
	PhaseItem       *eventpq.EventItem

	PendingPhase     int   // if paused, this is the next phase scheduled. is 0 by default. if 0, no pending phases
	PendingPhaseTime int64 // execute at active time; 0 means no pending phase

}

// TODO: hand/deck contents still aren't censored per client, only the Standoff bid fields are
func (match *Match) BroadcastPublishedState() {

	// send a separately censored copy to each client, since what's hidden depends on their team
	for key, client := range match.ByUser {
		team, ok := match.Whitelist[key]
		if !ok {
			continue
		}
		wrapper := struct {
			Type string      `json:"type"`
			Data interface{} `json:"data"`
		}{
			Type: "fullState",
			Data: state.CensorForTeam(match.PublishedState, int(team)),
		}
		gameStateJSON, err := json.Marshal(wrapper)
		if err != nil {
			continue
		}
		select {
		case client.LowPrio <- gameStateJSON:
		default:
			fmt.Println("dropped packet to", key)
		}
	}
}

// TODO: hand/deck contents still aren't censored per client, only the Standoff bid fields are
func (match *Match) SendPublishedState(ClientID string) {

	team, ok := match.Whitelist[ClientID]
	if !ok {
		return
	}

	// serialize current game state, censored for the requesting client's team
	wrapper := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: "fullState",
		Data: state.CensorForTeam(match.PublishedState, int(team)),
	}
	gameStateJSON, err := json.Marshal(wrapper)
	if err != nil {
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

// publishes the latest updates
func (match *Match) publish(wallTime int64, oldLogsSize int) {
	if len(match.EventLog) <= oldLogsSize {
		return
	}
	match.GameState.WallTimestamp = wallTime
	match.GameState.UpdateID = match.NextUpdateID - 1
	match.PublishedState = state.DeepCopyGameState(match.GameState)
	stateCopy := state.DeepCopyGameState(match.GameState)
	for _, log := range match.EventLog[oldLogsSize:] {
		match.DebugTrace = append(match.DebugTrace, debug.TraceEntry{Log: log, State: stateCopy})
	}
	//debug.WriteTrace(match.ID, match.DebugTrace)
	match.BroadcastLogs(oldLogsSize)
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
