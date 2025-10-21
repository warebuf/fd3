package matchmaking

type Room struct {
	id      string
	members map[*Client]struct{}
}
