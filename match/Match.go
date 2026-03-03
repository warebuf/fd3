package match

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/draw"
	"github.com/warebuf/fd3/engine/simulate"
)

type Match struct {
	UnitModels []engine.Model

	ID        string
	Whitelist map[string]uint // acts as the teamNo map and the whitelist
	ByUser    map[string]*Client

	GameState       *engine.GameState      // what the game state should be 200 ms from now
	PublishedState  *engine.GameState      // a copy of GameState that the players can read
	PQ              simulate.PriorityQueue // priority queue of chains/stacks
	NextEffectIndex uint64                 // assign an index to each action/effect
	ActionLog       []actions.Action       // a log of all events that have occurred

	NextRoundTime time.Time

	InputEventChan chan *InputEvent
	EventPQueue    EventPQ
	EventPQTimer   *time.Timer
	NextSeq        int
}

func (match *Match) DrawCard(team int, modelNo int) {
	t := time.Now()
	newState := draw.DrawUnit(match.UnitModels, match.GameStates[len(match.GameStates)-1], t, team, modelNo)
	match.GameStates = append(match.GameStates, newState)
	// some work here on incredmenting currentState on time
}

func (match *Match) CreateStructure(team int, HP int) {
	timestamp := time.Now()
	newState := engine.CopyGameState(timestamp, match.GameStates[len(match.GameStates)-1])
	newState.StructureHP = append(newState.StructureHP, HP)
	newState.StructureTeam = append(newState.StructureTeam, team)
	match.GameStates = append(match.GameStates, newState)
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

func (match *Match) BroadcastActions(acts []actions.Action) {
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

	curr := match.CurrentState

	for curr < len(match.GameStates) && time.Now().Add(200*time.Millisecond).After(match.GameStates[curr].Timestamp) {

		// serialize current game state
		state := match.GameStates[curr]
		wrapper := struct {
			Type string      `json:"type"`
			Data interface{} `json:"data"`
		}{
			Type: "fullState",
			Data: state,
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

		curr += 1

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
