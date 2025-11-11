package matchmaking

import (
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
}

func CreateMatchMaking(MatchMap map[string]*match.Match) *Matchmaker {
	matchMaker := &Matchmaker{
		queue:  orderedmap.New[string, *Ticket](),
		inGame: make(map[string]struct{}),
		Join:   make(chan *Ticket, 1024),
		Leave:  make(chan string, 1024),
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
				matchObj := m.createMatch(selected, MatchMap)

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

			// fmt.Println("m.queue.Len()", m.queue.Len())
		}
	}

}

func (m *Matchmaker) createMatch(tickets []*Ticket, MatchMap map[string]*match.Match) *match.Match {

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
		ID:        id,
		ByUser:    make(map[string]*match.Client),
		Whitelist: userIDs,

		GameStates:     make([]*engine.GameState, 0),
		CurrentState:   0,
		BroadcastTimer: nil,
	}

	state := engine.CreateGameState(time.Now(), match.CurrentState)
	match.GameStates = append(match.GameStates, state)

	MatchMap[id] = match

	return match

}
