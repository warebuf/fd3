package matchmaking

import (
	"sync"
)

type Waiting struct {
	UserID string
	Leave  chan bool
	Found  chan string // receives matchID
}

type Matchmaker struct {
	mu     sync.Mutex
	q      []Waiting           // FIFO queue of waiting users
	inGame map[string]struct{} // list of users in game
	join   chan Waiting        // new waiter arrives
	quit   chan string         // userID leaves/cancels
}
