package matchmaking

import (
	"context"
	"sync"
)

type Waiting struct {
	UserID string
	Leave  <-chan struct{}
	Found  chan string // receives matchID
}

type Matchmaker struct {
	mutex     sync.Mutex
	q      []*Waiting           // FIFO queue of waiting users
	inGame map[string]struct{} // list of users in game
	join   chan *Waiting        // new waiter arrives
	quit   chan string         // userID leaves/cancels
}

func CreateMatchMaking() {
	m := &Matchmaker{
		q: make([]string,0,1024),
		inGame: make(map[string]struct{}),
		join: make(chan *Waiting, 1024),
		quit: make(chan string, 1024)
	}
	go m.loop()
	return m
}

func (mm *Matchmaker) Join(ctx context.Context, UserID string) (out chan string) {
	out = make(chan string, 1)
	w := Waiting{
		UserID: UserID,
		Found:  out,
		Leave:  ctx.Done(),
	}
	mm.join <- w
	return out
}

func (mm *Matchmaker) loop() {

}
