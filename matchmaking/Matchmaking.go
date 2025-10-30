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
	mu     sync.Mutex
	q      []Waiting           // FIFO queue of waiting users
	inGame map[string]struct{} // list of users in game
	join   chan Waiting        // new waiter arrives
	quit   chan string         // userID leaves/cancels
}

func CreateMatchMaking() {

}

// maybe get rid of this and probably put it into SSE or something
func (mm *Matchmaker) Join(ctx context.Context, ticketID string) (out chan string) {
	out = make(chan string, 1)
	w := Waiting{
		UserID: ticketID,
		Found:  out,
		Leave:  ctx.Done(),
	}
	mm.join <- w
	return out
}

func (mm *Matchmaker) loop() {

}
