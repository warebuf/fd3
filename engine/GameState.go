package engine

type GameState struct {
	FieldUnits []uint // unitID
	DeadUnits  []uint // unitID

	// core
	HP []int

	// telemetry
	TimesAttacked  []uint
	TimesDefended  []uint
	DamageDealt    []uint
	DamageReceived []uint

	// counter
	NextID uint
}

// CreateMatchMaking initializes an empty game state with outer slices ready
func CreateMatchMaking() *GameState {
	gm := &GameState{
		FieldUnits:     make([]uint, 0),
		DeadUnits:      make([]uint, 0),
		HP:             make([]int, 0),
		TimesAttacked:  make([]uint, 0),
		TimesDefended:  make([]uint, 0),
		DamageDealt:    make([]uint, 0),
		DamageReceived: make([]uint, 0),
		NextID:         0,
	}
	return gm
}
