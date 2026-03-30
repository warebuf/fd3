package gameState

import (
	"time"
)

type GameState struct {
	GameStateID int
	Timestamp   time.Time // time the state is set for
	Phase       int       // what phase the game is in (0: pre-match, 1: draw hands, 2: play, 3: game end)

	//
	LastLogIndex int

	// state stuff
	Cards          []Card
	Permanents     []Permanent
	Mana           []int // per team
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
func CreateFirstGameState(t time.Time, cards []Card) *GameState {
	gm := &GameState{
		GameStateID:  0,
		Timestamp:    t,
		Phase:        0,
		LastLogIndex: 0,
		Cards:        cards,
	}
	return gm
}

func DeepCopyGameState(timestamp time.Time, prevState *GameState) *GameState {

	if prevState == nil {
		return nil
	}

	newState := &GameState{
		GameStateID: prevState.GameStateID + 1, // increment ID for new state
		Timestamp:   timestamp,
		Phase:       prevState.Phase,
	}

	// Cards
	newState.Cards = make([]Card, len(prevState.Cards))
	copy(newState.Cards, prevState.Cards)

	// Permanents
	newState.Permanents = make([]Permanent, len(prevState.Permanents))
	copy(newState.Permanents, prevState.Permanents)

	// Mana
	newState.Mana = make([]int, len(prevState.Mana))
	copy(newState.Mana, prevState.Mana)

	// ActiveTriggers
	newState.ActiveTriggers = make([]TriggerEffect, len(prevState.ActiveTriggers))
	copy(newState.ActiveTriggers, prevState.ActiveTriggers)

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
