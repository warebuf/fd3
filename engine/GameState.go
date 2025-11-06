package engine

type GameState struct {
	FieldUnits []Unit
	DeadUnits  []Unit
}

// CreateMatchMaking initializes an empty game state with outer slices ready
func CreateMatchMaking() *GameState {
	gm := &GameState{
		FieldUnits: make([]Unit, 0),
		DeadUnits:  make([]Unit, 0),
	}
	return gm
}
