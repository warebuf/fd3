package matchmaking

import (
	"math/rand"

	"github.com/warebuf/fd3/engine/constants"
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/state"
)

func DeckSetup(deckDefs [][]int, structureDefNos []int) *state.GameState {

	// TODO: validate the structure input

	// create structures; each team still submits one def number, so it fills all of their slots
	var structures [constants.NUM_TEAMS][constants.STRUCTURES_PER_TEAM]state.Structure
	for team, defNo := range structureDefNos {
		if team >= constants.NUM_TEAMS {
			break
		}
		def := defs.StructureDefs[defNo]
		for i := range structures[team] {
			structures[team][i] = state.Structure{
				DefNo:       defNo,
				Name:        def.Name,
				Description: def.Description,
				DrawCost:    def.DrawCost,
				HP:          def.BaseHP,
			}
		}
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

		Cards: []state.Card{},
		Deck:  make([][]int, numTeams),
		Hand:  make([][]int, numTeams),
		Field: make([][]int, numTeams),
		Grave: make([][]int, numTeams),
		Traps: make([][]int, numTeams),
		Equip: make([][]int, numTeams),
		Chain: []state.ChainItem{},
		Mana: func() []int {
			m := make([]int, numTeams)
			for i := range m {
				m[i] = constants.START_MANA
			}
			return m
		}(),
		Structures: structures,
		Ready:      make([]bool, numTeams),
		SkipChain:  make([]bool, numTeams),

		StandoffCardIndex: -1,
		PendingBidAmount: func() []int {
			b := make([]int, numTeams)
			for i := range b {
				b[i] = -1
			}
			return b
		}(),
	}
	nextCardIndex := 0
	for team := 0; team < len(deckDefs); team++ {
		defNos := deckDefs[team]

		gs.Deck[team] = make([]int, 0, len(defNos))
		gs.Hand[team] = []int{}
		gs.Field[team] = []int{}
		gs.Grave[team] = []int{}
		gs.Traps[team] = []int{}
		gs.Equip[team] = []int{}

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
				Bodyguard:   def.Bodyguard,

				Team:           team,
				Type:           def.Type,
				Zone:           state.ZoneDeck,
				IsToken:        def.IsToken,
				ChainEntry:     def.ChainEntry,
				Charges:        def.Charges,
				MaxCharges:     def.MaxCharges,
				PendingCharges: 0,
				AutoSummon:     0,
				EquipHostIndex: -1,

				AttackBonusPerSummon: def.AttackBonusPerSummon,
			}
			for i, ability := range def.OnPlay {
				card.OnPlay[i] = state.OnPlayAbility{
					Target:      -1,
					TargetFlags: ability.TargetFlags,
				}
			}
			for i, ability := range def.Active {
				card.Actives[i].Target = -1 // no target selected, even on an unused slot
				if ability.Effects[0].Type == 0 {
					continue // unused ability slot, leave the rest zero-valued so the client can tell
				}
				card.Actives[i] = state.ActiveAbility{
					Target:      -1,
					TargetFlags: ability.TargetFlags,
					Cons:        ability.Cons,
					Cost:        ability.Cost,
					CanAutoFire: ability.CanAutoFire,
					AutoAbility: false,
					ChainEntry:  ability.ChainEntry,
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

	if len(deck) > 100 {
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

		// tokens are spawned at runtime, not deck-buildable; retired DefNos are gaps in CardDefs
		if defs.CardDefs[cardNo].IsToken || defs.IsRetired(cardNo) {
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
