package matchmaking

import (
	"context"
	"fmt"
	"sync"
)

type Waiting struct {
	UserID string
	Leave  <-chan struct{}
	Found  chan string // receives matchID
}

type Matchmaker struct {
	mutex    sync.Mutex
	q        []*Waiting          // FIFO queue of waiting users
	inSystem map[string]struct{} // list of users in game
	join     chan *Waiting       // new waiter arrives
	quit     chan string         // userID leaves/cancels
}

func CreateMatchMaking() *Matchmaker {
	m := &Matchmaker{
		q:        make([]*Waiting, 0, 1024),
		inSystem: make(map[string]struct{}),
		join:     make(chan *Waiting, 1024),
		quit:     make(chan string, 1024),
	}
	go m.loop()
	return m
}

func (m *Matchmaker) Join(ctx context.Context, UserID string) (out chan string, err bool) {

	// check if userID is in the system, if so, do not join
	m.mutex.Lock()
	if _, exists := m.inSystem[UserID]; exists {
		return nil, false
	}
	m.inSystem[UserID] = struct{}{}
	m.mutex.Unlock()

	// create the waiting object, add to join channel
	out = make(chan string, 1)
	w := &Waiting{
		UserID: UserID,
		Found:  out,
		Leave:  ctx.Done(),
	}
	m.join <- w
	return out, true
}

func (m *Matchmaker) loop() {
	for {
		select {
		case w := <-m.join:
			m.mutex.Lock()

			// add the object to the queue

			// if possible, take the two latest

			// double check if they are both in the system

			// if they are not, clear and retry

			// if they are handle, the matchmaking like so
			w.Found <- "test"

			m.mutex.Unlock()
		case str := <-m.quit:
			fmt.Println("str", str)
		}
	}

}
