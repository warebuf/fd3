package sim

import (
	"math/rand"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/state"
)

// this could be in it's own package called match init
func CreateDecks(deckDefs [][]int, structureDefNos []int) *state.GameState {

	// create structures
	structures := make([]state.Structure, len(deckDefs))
	for i, defNo := range structureDefNos {
		structures[i] = state.Structure{DefNo: defNo, HP: defs.StructureDefs[defNo].BaseHP}
	}

	numTeams := len(deckDefs)
	timeBank := make([]int64, numTeams)
	for i := range timeBank {
		timeBank[i] = constants.PAUSE_BANK
	}

	gs := &state.GameState{
		UpdateID:        0,
		WallTimestamp:   0,
		Phase:           0,
		PausePhase:      2, // start in simulating state
		FullPauseOffset: 0,
		SimPauseOffset:  0,
		TeamPaused:      -1,
		Winner:          -1,
		TimeBank:        timeBank,

		Cards:         []state.Card{},
		Deck:          make([][]int, numTeams),
		Hand:          make([][]int, numTeams),
		Field:         make([][]int, numTeams),
		Grave:         make([][]int, numTeams),
		Continuous:    make([][]int, numTeams),
		ReactionStack: []state.ReactionItem{},
		Mana: func() []int {
			m := make([]int, numTeams)
			for i := range m {
				m[i] = constants.START_MANA
			}
			return m
		}(),
		Structures: structures,
		Ready:      make([]bool, numTeams),
	}
	nextCardIndex := 0
	for team := 0; team < len(deckDefs); team++ {
		defNos := deckDefs[team]

		gs.Deck[team] = make([]int, 0, len(defNos))
		gs.Hand[team] = []int{}
		gs.Field[team] = []int{}
		gs.Grave[team] = []int{}
		gs.Continuous[team] = []int{}

		for _, defNo := range defNos {
			def := defs.CardDefs[defNo]

			card := state.Card{
				DefNo:       defNo,
				Name:        def.Name,
				Description: def.Description,
				Cost:        def.Cost,
				HP:          def.BaseHP,
				Archetype:   def.Archetype,
				Taunt:       def.Taunt,

				Team:          team,
				Type:          def.Type,
				Zone:          state.ZoneDeck,
				AuraModifiers: []state.AuraModifier{},
			}
			for i, ability := range def.Active {
				card.Actives[i] = state.ActiveAbility{
					Charges:     ability.Charges,
					MaxCharges:  ability.MaxCharges,
					Target:      -1,
					TargetFlags: ability.TargetFlags,
					Cons:        ability.Cons,
					Cost:        ability.Cost,
				}
			}

			gs.Cards = append(gs.Cards, card)
			gs.Deck[team] = append(gs.Deck[team], nextCardIndex)

			nextCardIndex++
		}
	}

	return gs
}

func ShuffleDeck(gs *state.GameState, team int) {
	deck := gs.Deck[team]

	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

func IsValidSubmission(deck []int, startingHand []int, structNo int) bool {

	if len(deck) > 30 {
		return false
	}

	cardDefsLength := len(defs.CardDefs)
	counts := make([]int, cardDefsLength)

	for i := 0; i < len(deck); i++ {
		cardNo := deck[i]

		// bounds check
		if cardNo <= 0 || cardNo >= cardDefsLength { // rejecting 0, negative, and non-existing numbers
			return false
		}

		// count occurrences
		counts[cardNo]++

		// duplicate rule (example: max 3 copies)
		if counts[cardNo] > 3 {
			return false
		}
	}

	// validate starting hand: 0-5 cards, each DefNo must exist in the deck (consuming duplicates)
	if len(startingHand) > 5 {
		return false
	}
	handCounts := make([]int, cardDefsLength)
	for _, defNo := range startingHand {
		if defNo <= 0 || defNo >= cardDefsLength {
			return false
		}
		handCounts[defNo]++
		if handCounts[defNo] > counts[defNo] {
			return false
		}
	}

	// make sure structNo is available
	structDefsLength := len(defs.StructureDefs)
	if structNo <= 0 || structNo >= structDefsLength {
		return false
	}

	return true
}

func ArrangeStartingHand(gs *state.GameState, teamIndex int, startingHand []int) {
	if len(startingHand) == 0 {
		return
	}

	deck := make([]int, len(gs.Deck[teamIndex]))
	copy(deck, gs.Deck[teamIndex])

	chosen := []int{}
	for _, defNo := range startingHand {
		for i, cardIndex := range deck {
			if gs.Cards[cardIndex].DefNo == defNo {
				chosen = append(chosen, cardIndex)
				deck = append(deck[:i], deck[i+1:]...)
				break
			}
		}
	}

	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	gs.Deck[teamIndex] = append(chosen, deck...)
}
