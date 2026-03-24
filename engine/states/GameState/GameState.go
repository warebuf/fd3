package gameState

import (
	"time"
)

type GameState struct {
	GameStateID int
	Timestamp   time.Time // time the state is set for
	Round       int       // what round the state is on

	// state stuff
	Cards          []Card
	Permanents     []Permanent
	Mana           []int
	ActiveTriggers []TriggerEffect
}

type Permanent struct {
	PermanentID int
	SourceCard  int
	Type        int // structure, unit, continuous trap, continuous spell
	HP          int
	Team        int
	Master      bool // can only be 1 master unit
}

// CreateGameState initializes an empty game state with outer slices ready
func CreateFirstGameState(t time.Time, GSID int) *GameState {
	gm := &GameState{
		GameStateID: GSID,
		Timestamp:   t,
		Round:       0,
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
		Round:       prevState.Round,
	}

	newState.Mana = copyIntSlice(prevState.Mana)

	return newState
}

// helpers

func copyIntSlice(src []int) []int {
	if src == nil {
		return nil
	}
	dst := make([]int, len(src))
	copy(dst, src)
	return dst
}
