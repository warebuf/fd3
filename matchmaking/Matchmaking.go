package matchmaking

import (
	"sync"

	"github.com/google/uuid"
	orderedmap "github.com/wk8/go-ordered-map/v2"
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

	inGame   map[string]struct{} // list of users in game
	MatchMap map[string]*Match
}

func CreateMatchMaking() *Matchmaker {
	m := &Matchmaker{
		queue:    orderedmap.New[string, *Ticket](),
		inGame:   make(map[string]struct{}),
		Join:     make(chan *Ticket, 1024),
		Leave:    make(chan string, 1024),
		MatchMap: make(map[string]*Match),
	}
	go m.loop()
	return m
}

func (m *Matchmaker) loop() {
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
				match := m.createMatch(selected)
				for i := 0; i < len(selected); i++ {
					select {
					case selected[i].Found <- match.ID:
					default:
					}
					select {
					case selected[i].Found <- match.ID:
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

func (m *Matchmaker) createMatch(tickets []*Ticket) *Match {

	// create matchID
	id := uuid.NewString()
	for {
		if _, exists := m.MatchMap[id]; !exists {
			break
		}
		id = uuid.NewString()
	}

	// create whitelist userIDs
	userIDs := make(map[string]struct{})
	for i := 0; i < len(tickets); i++ {
		userIDs[tickets[i].TicketID] = struct{}{}
	}

	// create match object
	match := &Match{
		ID:        id,
		byUser:    make(map[string]*Client),
		Whitelist: userIDs,
	}
	return match

}
