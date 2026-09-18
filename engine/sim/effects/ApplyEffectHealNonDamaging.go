package effects

import (
	"math"

	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// gives amount HP to every unit on the field, on both sides, whose CardDef has no damaging ability;
// each affected unit gets its own HealLog, same as a single-target heal
func ApplyEffectHealNonDamaging(gs *state.GameState, logList *[]any, UpdateID *int, eff state.Effect, executeAtWallTime int64) {
	amount := eff.B

	for team := 0; team < len(gs.Field); team++ {
		for _, cardIndex := range gs.Field[team] {
			card := &gs.Cards[cardIndex]
			if card.Type != state.CardTypeUnit || defs.IsDamaging(card.DefNo) {
				continue
			}

			if amount > math.MaxInt-card.HP {
				card.HP = math.MaxInt
			} else {
				card.HP += amount
			}

			*logList = append(*logList, logs.HealLog{
				UpdateID:        *UpdateID,
				WallTimestamp:   executeAtWallTime,
				Type:            logs.Heal,
				HealerIndex:     cardIndex,
				TargetCardIndex: cardIndex,
				Amount:          amount,
			})
			*UpdateID++
		}
	}
}
