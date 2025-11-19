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

TODO: experiment with different latencies (200 ms - 2 s)
TODO: add concurrency data structures for needed resources
*/

type Match struct {
	ID        string
	Whitelist map[string]uint // acts as the teamNo map and the whitelist
	ByUser    map[string]*Client

	GameStates   []*engine.GameState
	CurrentState uint

	ChangedHeap    chan bool
	BroadcastTimer *time.Timer
}

func (match *Match) SummonUnit(team uint, modelNo uint) {
	// set a summon for 200 ms
	t := time.Now().Add(200 * time.Millisecond)
	newState := summon.BasicSummon(match.GameStates[len(match.GameStates)-1], t, team, modelNo)
	match.GameStates = append(match.GameStates, newState)
	// maybe some work here around incrementing currentState on time
}

func (match *Match) DrawCard(team uint, modelNo uint) {
	t := time.Now().Add(200 * time.Millisecond)
	newState := draw.DrawUnit(match.GameStates[len(match.GameStates)-1], t, team, modelNo)
	match.GameStates = append(match.GameStates, newState)
	// some work here on incredmenting currentState on time
}

// pseudocode:
// func (match *Match) Loop() {
//     timer := time.NewTimer(time.Hour * 8760) // dummy long timer
//     timer.Stop()                      // disable until we set it
//     var timerC <-chan time.Time       // nil until active

//     for {
//         select {
// 		   case <-match.ChangedHeap:
//            // no events left
//            if match.CurrentState+1 >= len(match.GameStates) {
//               timer.Stop()
//               timerC = nil
//               continue
//            }
//
//            // Skip all expired timestamps
//            for {
//               nextTs := match.GameStates[match.CurrentState+1].Timestamp
//               if time.Until(nextTs) > 0 {
//                  break
//               }
//               match.CurrentState++
//               if match.CurrentState+1 >= len(match.GameStates) {
//                  break
//               }
//            }
//
//            // Safely stop and drain timer
//            if !timer.Stop() {
//               select {
//               case <-timer.C:
//               default:
//               }
//            }
//
//            // If we still have events, schedule next
//            if match.CurrentState+1 < len(match.GameStates) {
//               d := time.Until(match.GameStates[match.CurrentState+1])
//               timer.Reset(d) // the timer will fire immediately if d is already expired (ie <= 0)
//               timerC = timer.C
//            } else { // no future events
//               timer.Stop()
//               timerC = nil
//            }
//             
//         case <-timerC:
//             
//             // i mean for now this code will simply increment to the next timestamp
//             // but going forward, it seems like we have to predict the next gamestate here
//
//             pseudocode:
//             increment the state until we get a non-expired timestamp
//             recalculate the next predicted state vs the next inputState
//             reconcile the eventSequencer
//             schedule the next timer
//             
//         }
//     }
// }

// simply increments currentState, not completely functional afaik
func (match *Match) Loop() {

	var timerChannel <-chan time.Time // set to nil on init

	for {
		select {
		case isRunning := <-match.ChangedTimer:
			fmt.Println("changedTimer")
			if isRunning {
				fmt.Println("match.BroadcastTimer", match.BroadcastTimer)
				timerChannel = match.BroadcastTimer.C
			} else {
				fmt.Println("nil")
				timerChannel = nil
			}
		case <-timerChannel:
			fmt.Println("timer ran", match.CurrentState, len(match.GameStates))
		}
	}
}

// TODO: needs to be censored per client
func (match *Match) BroadcastGameState() {

	// serialize current game state
	state := match.GameStates[match.CurrentState]
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

	for int(curr) < len(match.GameStates) && time.Now().Add(200*time.Millisecond).After(match.GameStates[curr].Timestamp) {

		fmt.Println("SendGameState curr", curr)

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

func (match *Match) Print() {
	for i := 0; i < len(match.GameStates); i++ {
		fmt.Printf("%d %+v\n", match.GameStates[i].Timestamp.UnixMilli(), match.GameStates[i])
	}
}
