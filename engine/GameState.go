package engine

import "time"

type GameState struct {
	Timestamp time.Time // what time this is set for

	// state sets: array of unitIDs
	BenchUnits []uint
	FieldUnits []uint
	DeadUnits  []uint

	// team sets: index (unitID) ->  team index
	Team []uint

	// model set
	ModelNo []uint

	// core unit stats: index (unitID) ->  HP
	HP []int

	// telemetry
	TimesAttacked  []uint // index (unitID) -> val
	TimesDefended  []uint
	DamageDealt    []uint
	DamageReceived []uint

	// counter
	NextID uint
}

// CreateGameState initializes an empty game state with outer slices ready
func CreateGameState() *GameState {
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
