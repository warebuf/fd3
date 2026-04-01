package gameState

import (
	"time"
)

type GameState struct {
	GameStateID int
	Timestamp   time.Time // time the state is set for
	Phase       int       // what phase the game is in (0: pre-match, 1: draw hands, 2: play, 3: game end)

	// card states
	Cards []Card
	Deck  [][]int
	Hand  [][]int // front of arr is the top
	Field [][]int
	Grave [][]int

	// other states
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

func DeepCopyGameState(prevState *GameState) *GameState {
	newState := &GameState{
		GameStateID: prevState.GameStateID, // increment ID for new state
		Timestamp:   prevState.Timestamp,
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
