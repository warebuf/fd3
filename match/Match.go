package match

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/engine/draw"
	"github.com/warebuf/fd3/engine/summon"
)

/*
note:
the only flaw in this system is that if a client has
a lag spike or has bad internet, they will not receive
the new game states in time

if they have a lag spike, not my problem, they missed the event
and will have to request the game state to get caught up

if they are consistently having issues, let them know
*/

type Match struct {
	UnitModels []engine.Model

	ID        string
	Whitelist map[string]uint // acts as the teamNo map and the whitelist
	ByUser    map[string]*Client

	GameStates   []*engine.GameState
	CurrentState int

	InputEventChan chan *InputEvent
	EventPQueue    EventPQ
	EventPQTimer   *time.Timer
	NextSeq        int
}

func (match *Match) SummonUnit(team int, modelNo int) {
	t := time.Now()
	newState := summon.BasicSummon(match.UnitModels, match.GameStates[len(match.GameStates)-1], t, team, modelNo)
	match.GameStates = append(match.GameStates, newState)
	// maybe some work here around incrementing currentState on time
}

func (match *Match) DrawCard(team int, modelNo int) {
	t := time.Now()
	newState := draw.DrawUnit(match.UnitModels, match.GameStates[len(match.GameStates)-1], t, team, modelNo)
	match.GameStates = append(match.GameStates, newState)
	// some work here on incredmenting currentState on time
}

// TODO: needs to be censored per client
func (match *Match) BroadcastGameState(state *engine.GameState) {

	// serialize game state
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
