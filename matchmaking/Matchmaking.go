package matchmaking

import (
	"container/heap"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	orderedmap "github.com/wk8/go-ordered-map/v2"

	"github.com/warebuf/fd3/engine/sim"
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match"
	"github.com/warebuf/fd3/match/eventpq"
)

type Ticket struct {
	TicketID         string
	Deck             []int
	StructureDefNo   int
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
		queue: orderedmap.New[string, *Ticket](),
		Join:  make(chan *Ticket, 1024),
		Leave: make(chan string, 1024),

		inGame: make(map[string]struct{}),
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
				select { // non-blocking
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

				matchObj := CreateMatch(selected, MatchMap) // create match

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

func CreateMatch(tickets []*Ticket, MatchMap map[string]*match.Match) *match.Match {

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

	// creating the first state, instantiating the card objects
	var deckDefs [][]int
	var structureDefNos []int
	for i := 0; i < len(tickets); i++ {
		deckDefs = append(deckDefs, tickets[i].Deck)
		structureDefNos = append(structureDefNos, tickets[i].StructureDefNo)
	}
	gs := sim.CreateDecks(deckDefs, structureDefNos)

	// shuffle the decks
	for i := 0; i < len(gs.Deck); i++ {
		sim.ShuffleDeck(gs, i)
	}

	// create match object
	m := &match.Match{
		ID:        id,
		ByUser:    make(map[string]*match.Client),
		Whitelist: userIDs,

		GameState:      gs,
		PublishedState: state.DeepCopyGameState(gs),
		EventLog:       []any{},
		NextUpdateID:   1, // starts as 1 because the GameState already is UpdateID 0

		InputEventChan: make(chan *eventpq.InputEvent, 1024),
		EventPQTimer:   nil,
	}
	heap.Init(&m.EventPQueue) // init event handling PQ

	match.PushPhase(m, time.Now().UnixMilli()+500, 1) // timer for starting the game, could also broadcast the time here

	go m.EventScheduler() // on init, check the game state, schedule if first if possible

	MatchMap[id] = m

	return m

}
