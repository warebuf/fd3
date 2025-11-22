package engine

import (
	"time"

	copyslice "github.com/warebuf/fd3/utility/copySlice"
)

type GameState struct {
	// time the state is set for
	GameStateID int
	Timestamp   time.Time

	// team sets: index (unitID) ->  team index
	Team []int
	Zone []int // 0: Bench, 1: Field, 2: Dead

	// model set
	ModelNo []int

	// core unit stats: index (unitID) ->  HP
	HP []int

	// abilities
	Abilities       []int // abilities are flattened for each unit
	AbilityStartIdx []int // the starting index in Abilities[] for this unit
	AbilityCount    []int // how many abilities belong to each unit

	AbilityCharge   []int64
	AbilityCooldown []int64

	// counter
	NextUnitID uint
}

type Telemetry struct {
	TimesAttacked  []int64 // index (unitID) -> val
	TimesDefended  []int64
	DamageDealt    []int64
	DamageReceived []int64
}

// CreateGameState initializes an empty game state with outer slices ready
func CreateGameState(t time.Time, GSID int) *GameState {
	gm := &GameState{
		GameStateID: GSID,
		Timestamp:   t,
		Team:        make([]int, 0),
		Zone:        make([]int, 0),
		ModelNo:     make([]int, 0),
		HP:          make([]int, 0),
		NextUnitID:  0,
	}
	return gm
}

func CopyGameState(timestamp time.Time, prevState *GameState) *GameState {
	if prevState == nil {
		return nil
	}

	newState := &GameState{
		GameStateID: prevState.GameStateID + 1, // increment ID for new state
		Timestamp:   timestamp,
		NextUnitID:  prevState.NextUnitID,
	}

	newState.Team = copyslice.CopySliceInt(prevState.Team)
	newState.Zone = copyslice.CopySliceInt(prevState.Zone)
	newState.ModelNo = copyslice.CopySliceInt(prevState.ModelNo)
	newState.HP = copyslice.CopySliceInt(prevState.HP)
	return newState
}
