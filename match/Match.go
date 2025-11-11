package match

import (
	"fmt"
	"time"

	"github.com/warebuf/fd3/engine"
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
*/

type Match struct {
	ID        string
	Whitelist map[string]uint // acts as the teamNo map and the whitelist
	ByUser    map[string]*Client

	GameStates   []*engine.GameState
	CurrentState uint

	ChangedTimer   chan bool
	BroadcastTimer *time.Timer
}

func (match *Match) AddTestAttacker() {
	// set a summon for 500 ms
	t := time.Now().Add(500 * time.Millisecond)
	newState := summon.BasicSummon(match.GameStates[match.CurrentState], t)
	match.GameStates = append(match.GameStates, newState)

	// set the broadcast timer for 300 ms
	target := time.Until(t.Add(-200 * time.Millisecond))
	match.BroadcastTimer = time.NewTimer(target)
	match.ChangedTimer <- true
}

func (match *Match) Loop() {

	var timerChannel <-chan time.Time // set to nil on init

	for {
		select {
		case isRunning := <-match.ChangedTimer:
			if isRunning {
				timerChannel = match.BroadcastTimer.C
			} else {
				timerChannel = nil
			}
		case <-timerChannel:
			fmt.Println("timer ran")
		}
	}
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
