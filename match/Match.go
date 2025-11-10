package match

import (
	"fmt"
	"sync"
	"time"

	"github.com/warebuf/fd3/engine"
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
*/

type Match struct {
	ID        string
	Whitelist map[string]uint // acts as the teamNo map and the whitelist
	ByUser    map[string]*Client

	Mutex        sync.RWMutex
	GameStates   []*engine.GameState
	CurrentState uint
	ticker       time.Ticker
}

func (match *Match) addATestAttacker() {
	// add a summon 300 ms from now
	//   - create the game state object
	//   - set the timer
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
