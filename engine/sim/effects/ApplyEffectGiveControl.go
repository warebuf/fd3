package effects

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/destruction"
	"github.com/warebuf/fd3/engine/state"
)

// hands a friendly field unit to the other team. it keeps its HP, charges and scheduled ticks,
// but Team follows it, so its ticks now fire at its former controller and its grave is theirs
func ApplyEffectGiveControl(gs *state.GameState, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	fromTeam := gs.Cards[sourceCardIndex].Team
	targetIndex := eff.A

	if targetIndex < 0 || targetIndex >= len(gs.Cards) {
		return // preselected target was never set, ability fizzles
	}
	card := &gs.Cards[targetIndex]
	if card.Zone != state.ZoneField || card.Team != fromTeam {
		return // target left the field or already changed hands, ability fizzles
	}

	moved := false
	for i, ci := range gs.Field[fromTeam] {
		if ci == targetIndex {
			gs.Field[fromTeam] = append(gs.Field[fromTeam][:i], gs.Field[fromTeam][i+1:]...)
			moved = true
			break
		}
	}
	if !moved {
		return
	}

	toTeam := 1 - fromTeam
	gs.Field[toTeam] = append(gs.Field[toTeam], targetIndex)
	card.Team = toTeam

	// the old controller's targeting and auto-fire choices don't carry over to the new one
	for i := range card.Actives {
		card.Actives[i].Target = -1
		card.Actives[i].TargetStructure = -1
		card.Actives[i].AutoAbility = false
	}

	// preselects aimed at it were made against the side it used to be on
	destruction.ClearPreselectsTargeting(gs, targetIndex)

	*logList = append(*logList, logs.GiveControlLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.GiveControl,
		CardIndex:     targetIndex,
		FromTeamIndex: fromTeam,
		ToTeamIndex:   toTeam,
	})
	*UpdateID++
}
