package matchmaking

type Waiter struct {
	UserID string
	Leave chan bool
	Found chan string // receives matchID
}

type Matchmaker struct {
	mu sync.Mutex
	q []Waiter                   // FIFO queue of waiting users
	inGame map[string]struct{}            // list of users in game
	join chan Waiter                // new waiter arrives
	quit chan string                // userID leaves/cancels
}