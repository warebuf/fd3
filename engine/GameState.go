package engine

type GameState struct {
	FieldUnits []uint // unitID
	DeadUnits  []uint // unitID

	// core

	// telemetry
	TimesAttacked []uint
	TimesDefended []uint
	DamageDealt   []uint
}

// CreateMatchMaking initializes an empty game state with outer slices ready
func CreateMatchMaking() *GameState {
	gm := &GameState{
		FieldUnits: make([]Unit, 0),
		DeadUnits:  make([]Unit, 0),
	}
	return gm
}
