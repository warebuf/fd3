package effects

import (
	"math/rand"

	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/destruction"
	"github.com/warebuf/fd3/engine/state"
)

// A: the activator's own chosen unit (from Card.Actives[abilityIndex].Target). Picks a random unit
// off the opponent's field to stand in for "the opponent's choice" (the engine has no bilateral
// target-choice mechanic yet), swaps control of the two so each now belongs to the other side, then
// has both immediately attack using the same random/taunt-aware targeting a normal attack uses.
func ApplyEffectMutualDestruction(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	myIndex := eff.A
	if myIndex < 0 || myIndex >= len(gs.Cards) {
		return // preselected target was never set, ability fizzles
	}
	activatorTeam := gs.Cards[sourceCardIndex].Team
	myCard := gs.Cards[myIndex]
	if myCard.Zone != state.ZoneField || myCard.Team != activatorTeam {
		return // target left the field or already changed hands, ability fizzles
	}

	opponentTeam := 1 - activatorTeam
	opponentField := gs.Field[opponentTeam]
	if len(opponentField) == 0 {
		return // nothing on the opponent's side to exchange, ability fizzles
	}
	theirIndex := opponentField[rand.Intn(len(opponentField))]

	swapFieldControl(gs, logList, UpdateID, myIndex, activatorTeam, opponentTeam, executeAtWallTime)
	swapFieldControl(gs, logList, UpdateID, theirIndex, opponentTeam, activatorTeam, executeAtWallTime)

	attackEff := state.Effect{Type: state.EffectAttack, A: 0, B: eff.B}
	ApplyEffectAttack(gs, pq, logList, UpdateID, attackEff, myIndex, -1, executeAtWallTime)
	ApplyEffectAttack(gs, pq, logList, UpdateID, attackEff, theirIndex, -1, executeAtWallTime)
}

// moves cardIndex from fromTeam's field to toTeam's, resetting its targeting/auto-fire choices;
// same bookkeeping as ApplyEffectGiveControl, factored out so both swap directions share it
func swapFieldControl(gs *state.GameState, logList *[]any, UpdateID *int, cardIndex int, fromTeam int, toTeam int, executeAtWallTime int64) {
	for i, ci := range gs.Field[fromTeam] {
		if ci == cardIndex {
			gs.Field[fromTeam] = append(gs.Field[fromTeam][:i], gs.Field[fromTeam][i+1:]...)
			break
		}
	}
	gs.Field[toTeam] = append(gs.Field[toTeam], cardIndex)

	card := &gs.Cards[cardIndex]
	card.Team = toTeam
	for i := range card.Actives {
		card.Actives[i].Target = -1
		card.Actives[i].TargetStructure = -1
		card.Actives[i].AutoAbility = false
	}
	destruction.ClearPreselectsTargeting(gs, cardIndex)

	*logList = append(*logList, logs.GiveControlLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.GiveControl,
		CardIndex:     cardIndex,
		FromTeamIndex: fromTeam,
		ToTeamIndex:   toTeam,
	})
	*UpdateID++
}
