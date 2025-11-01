package matchmaking

import (
	"sync"
)

type Waiting struct {
	UserID string
	Leave  <-chan struct{}
	Found  chan string // receives matchID
}

type Matchmaker struct {
	mutex    sync.Mutex
	q        []string            // FIFO queue of waiting users
	inSystem map[string]*Waiting // list of users in game
	Join     chan *Waiting       // new waiter arrives
	Leave    chan string
}

func CreateMatchMaking() *Matchmaker {
	m := &Matchmaker{
		q:        make([]string, 0, 1024),
		inSystem: make(map[string]*Waiting),
		Join:     make(chan *Waiting, 1024),
		Leave:    make(chan string, 1024),
	}
	go m.loop()
	return m
}

func (m *Matchmaker) loop() {
	for {
		select {
		case w := <-m.Join:
			m.mutex.Lock()

			// check if ticketID is in the system
			// if it is, ignore

			// if it is not, add to the queue/system

			// take the two latest that are in the system, discard others

			// matchmake
			w.Found <- "test"

			m.mutex.Unlock()
		}
	}

}
