package matchmaking

import (
	"sync"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// lots of test cases we need to check

type Ticket struct {
	TicketID string
	Leave    <-chan struct{}
	Found    chan string   // receives matchID
	Reject   chan struct{} // if UserID is in the system, reject
}

type Matchmaker struct {
	mutex  sync.Mutex
	queue  *orderedmap.OrderedMap[string, *Ticket] // users waiting for a game, need to use a linked list
	inGame map[string]struct{}                     // list of users in game
	Join   chan *Ticket                            // new waiter arrives
	Leave  chan string
}

func CreateMatchMaking() *Matchmaker {
	m := &Matchmaker{
		queue:  orderedmap.New[string, *Ticket](),
		inGame: make(map[string]struct{}),
		Join:   make(chan *Ticket, 1024),
		Leave:  make(chan string, 1024),
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
				ticket.Reject <- struct{}{}
			}

			// if ticketID is in queue, return the SSE connection
			if _, exists := m.queue.Get(ticket.TicketID); exists {
				ticket.Reject <- struct{}{}
			}

			// add the ticket to the queue
			m.queue.Set(ticket.TicketID, ticket)

			// terate from newest to oldest, take the two latest that are in the system
			var first, second *Ticket
			for current := m.queue.Newest(); current != nil && (first == nil || second == nil); {

				// non-blocking
				select {
				case <-current.Value.Leave: // user is not connected, invalidate ticket
					prev := current.Prev()
					m.queue.Delete(current.Key)
					current = prev
					// best-effort notify reject without blocking
					select {
					case current.Value.Reject <- struct{}{}:
					default:
					}
				default: // save the ticket
					if first == nil {
						first = current.Value
					} else if second == nil {
						second = current.Value
					}
					current = current.Prev()
				}

			}

			// matchmake
			if first != nil && second != nil {
				first.Found <- "test"
				second.Found <- "test"
				m.queue.Delete(first.TicketID)
				m.queue.Delete(second.TicketID)
				m.inGame[first.TicketID] = struct{}{}
				m.inGame[second.TicketID] = struct{}{}
			}

			m.mutex.Unlock()
		}
	}

}
