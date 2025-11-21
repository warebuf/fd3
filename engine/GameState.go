package engine

import (
	"time"

	copyslice "github.com/warebuf/fd3/utility/copySlice"
)

type GameState struct {
	// time the state is set for
	GameStateID uint
	Timestamp   time.Time

	// team sets: index (unitID) ->  team index
	Team []uint
	Zone []uint // 0: Bench, 1: Field, 2: Dead

	// model set
	ModelNo []uint

	// core unit stats: index (unitID) ->  HP
	HP []int

	// abilities
	Abilities       []uint // abilities are flattened for each unit
	AbilityStartIdx []uint // the starting index in Abilities[] for this unit
	AbilityCount    []uint // how many abilities belong to each unit

	// counter
	NextUnitID uint
}

type Telemetry struct {
	TimesAttacked  []uint // index (unitID) -> val
	TimesDefended  []uint
	DamageDealt    []uint
	DamageReceived []uint
}

// CreateGameState initializes an empty game state with outer slices ready
func CreateGameState(t time.Time, GSID uint) *GameState {
	gm := &GameState{
		GameStateID: GSID,
		Timestamp:   t,
		Team:        make([]uint, 0),
		Zone:        make([]uint, 0),
		ModelNo:     make([]uint, 0),
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

	newState.Team = copyslice.CopySliceUint(prevState.Team)
	newState.Zone = copyslice.CopySliceUint(prevState.Zone)
	newState.ModelNo = copyslice.CopySliceUint(prevState.ModelNo)
	newState.HP = copyslice.CopySliceInt(prevState.HP)
	return newState
}
