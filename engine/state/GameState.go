package state

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
		GameStateID: prevState.GameStateID,
		Timestamp:   prevState.Timestamp,
		Phase:       prevState.Phase,
	}

	// Cards
	newState.Cards = make([]Card, len(prevState.Cards))
	copy(newState.Cards, prevState.Cards)

	// Deck
	newState.Deck = make([][]int, len(prevState.Deck))
	for i := range prevState.Deck {
		newState.Deck[i] = make([]int, len(prevState.Deck[i]))
		copy(newState.Deck[i], prevState.Deck[i])
	}

	// Hand
	newState.Hand = make([][]int, len(prevState.Hand))
	for i := range prevState.Hand {
		newState.Hand[i] = make([]int, len(prevState.Hand[i]))
		copy(newState.Hand[i], prevState.Hand[i])
	}

	// Field
	newState.Field = make([][]int, len(prevState.Field))
	for i := range prevState.Field {
		newState.Field[i] = make([]int, len(prevState.Field[i]))
		copy(newState.Field[i], prevState.Field[i])
	}

	// Grave
	newState.Grave = make([][]int, len(prevState.Grave))
	for i := range prevState.Grave {
		newState.Grave[i] = make([]int, len(prevState.Grave[i]))
		copy(newState.Grave[i], prevState.Grave[i])
	}

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
