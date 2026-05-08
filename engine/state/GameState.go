package state

import (
	"time"
)

type GameState struct {
	UpdateID  int
	Timestamp time.Time // time the state is set for

	Phase int // 0: pre-match, 1: draw, 2: board setup, 3: reveal, 4: play, 5: game end

	// card states
	Cards []Card
	Deck  [][]int
	Hand  [][]int // front of arr is the top
	Field [][]int
	Grave [][]int

	// reaction state
	ReactionPQ PriorityQueue

	// other states
	Mana        []int       // per team
	Structures  []Structure // per team
	StagedCards [][]int     // per team, cards hidden during board setup phase
	Ready       []bool      // per team, whether the player has readied up
}

func DeepCopyGameState(prevState *GameState) *GameState {
	newState := &GameState{
		UpdateID:  prevState.UpdateID,
		Timestamp: prevState.Timestamp,
		Phase:     prevState.Phase,
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

	// Mana
	newState.Mana = make([]int, len(prevState.Mana))
	copy(newState.Mana, prevState.Mana)

	// Structures
	newState.Structures = make([]Structure, len(prevState.Structures))
	copy(newState.Structures, prevState.Structures)

	// StagedCards
	newState.StagedCards = make([][]int, len(prevState.StagedCards))
	for i := range prevState.StagedCards {
		newState.StagedCards[i] = make([]int, len(prevState.StagedCards[i]))
		copy(newState.StagedCards[i], prevState.StagedCards[i])
	}

	// Ready
	newState.Ready = make([]bool, len(prevState.Ready))
	copy(newState.Ready, prevState.Ready)

	// ReactionPQ
	newPQ := PriorityQueue{
		total: prevState.ReactionPQ.total,
		Items: make([]*Chain, len(prevState.ReactionPQ.Items)),
	}
	for i, s := range prevState.ReactionPQ.Items {
		newChain := &Chain{
			ChainID:     s.ChainID,
			EffectIndex: s.EffectIndex,
			PopTime:     s.PopTime,
			index:       s.index,
			ChainSteps:  make([]ChainStep, len(s.ChainSteps)),
		}
		for j, se := range s.ChainSteps {
			newStep := ChainStep{
				Type:            se.Type,
				CardID:          se.CardID,
				StartTime:       se.StartTime,
				PopTime:         se.PopTime,
				OccurTime:       se.OccurTime,
				Effects:         make([]Effect, len(se.Effects)),
				RecurringPeriod: se.RecurringPeriod,
				Persistent:      se.Persistent,
			}
			copy(newStep.Effects, se.Effects)
			newChain.ChainSteps[j] = newStep
		}
		newPQ.Items[i] = newChain
	}
	newState.ReactionPQ = newPQ

	return newState
}
