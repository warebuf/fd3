package matchmaking

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	orderedmap "github.com/wk8/go-ordered-map/v2"

	"github.com/warebuf/fd3/engine"
	"github.com/warebuf/fd3/match"
)

type Ticket struct {
	TicketID         string
	ClosedConnection <-chan struct{}
	Found            chan string   // receives matchID
	RejectedTicket   chan struct{} // if UserID is in the system, reject
}

type Matchmaker struct {
	mutex sync.Mutex
	queue *orderedmap.OrderedMap[string, *Ticket] // users waiting for a game, need to use a linked list
	Join  chan *Ticket                            // new waiter arrives
	Leave chan string

	inGame map[string]struct{} // list of users in game

	UnitModels []engine.Model
}

func CreateMatchMaking(MatchMap map[string]*match.Match, unitModels []engine.Model) *Matchmaker {
	matchMaker := &Matchmaker{
		queue: orderedmap.New[string, *Ticket](),
		Join:  make(chan *Ticket, 1024),
		Leave: make(chan string, 1024),

		inGame: make(map[string]struct{}),

		UnitModels: unitModels,
	}
	go matchMaker.loop(MatchMap)
	return matchMaker
}

func (m *Matchmaker) loop(MatchMap map[string]*match.Match) {
	for {
		select {
		case ticket := <-m.Join:
			m.mutex.Lock()

			// if ticketID is in game, return the SSE connection
			if _, exists := m.inGame[ticket.TicketID]; exists {
				select {
				case ticket.RejectedTicket <- struct{}{}:
				default:
				} // non-blocking reject
				continue
			}

			// if ticketID is in queue, return the SSE connection
			if _, exists := m.queue.Get(ticket.TicketID); exists {
				select {
				case ticket.RejectedTicket <- struct{}{}:
				default:
				} // non-blocking reject
				continue
			}

			// add the ticket to the queue
			m.queue.Set(ticket.TicketID, ticket)

			// iterate from newest to oldest, take the two latest that are in the system
			var selected []*Ticket
			for current := m.queue.Newest(); current != nil && len(selected) < 2; {

				// non-blocking
				select {
				case <-current.Value.ClosedConnection: // user is not connected, invalidate ticket
					prev := current.Prev()
					m.queue.Delete(current.Key)
					select {
					case current.Value.RejectedTicket <- struct{}{}:
					default:
					} // make sure the ticket is closed
					current = prev
				default: // save the ticket
					selected = append(selected, current.Value)
					current = current.Prev()
				}

			}

			// matchmake
			if len(selected) >= 2 {

				// create match
				matchObj := CreateMatch(selected, MatchMap, m.UnitModels)

				// send match to all users
				for i := 0; i < len(selected); i++ {
					select {
					case selected[i].Found <- matchObj.ID:
					default:
					}
					select {
					case selected[i].Found <- matchObj.ID:
					default:
					}
					m.queue.Delete(selected[i].TicketID)
					m.queue.Delete(selected[i].TicketID)
					m.inGame[selected[i].TicketID] = struct{}{}
					m.inGame[selected[i].TicketID] = struct{}{}
				}
			}

			m.mutex.Unlock()

		case <-m.Leave:
			fmt.Println("havent programmed this yet, just in to remove syntax warning")

		}
	}

}

func CreateMatch(tickets []*Ticket, MatchMap map[string]*match.Match, unitModels []engine.Model) *match.Match {

	// create matchID
	id := uuid.NewString()
	for {
		if _, exists := MatchMap[id]; !exists {
			break
		}
		id = uuid.NewString()
	}

	// create whitelist userIDs
	userIDs := make(map[string]uint)
	for i := 0; i < len(tickets); i++ {
		userIDs[tickets[i].TicketID] = uint(i)
	}

	// create match object
	match := &match.Match{
		UnitModels: unitModels,

		ID:        id,
		ByUser:    make(map[string]*match.Client),
		Whitelist: userIDs,

		GameStates:        make([]*engine.GameState, 0),
		CurrentState:      0,
		GameStatesChanged: make(chan struct{}),
		CurrentStateTimer: nil,

		Events:         make([]match.Event, 0),
		InputEventChan: make(chan *match.InputEvent, 1024),
	}

	state := engine.CreateFirstGameState(time.Now(), match.CurrentState)
	match.GameStates = append(match.GameStates, state)

	match.DrawCard(0, 0)
	match.DrawCard(0, 1)
	match.DrawCard(1, 0)
	match.DrawCard(1, 1)

	go match.UpdateCurrentState()
	go match.EventScheduler() // on init, check the game state, schedule if first if possible

	MatchMap[id] = match

	return match

}
