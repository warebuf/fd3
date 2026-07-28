package effects

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/destruction"
	"github.com/warebuf/fd3/engine/sim/schedule"
	"github.com/warebuf/fd3/engine/state"
)

func ApplyEffectConvertToTokens(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, eff state.Effect, sourceCardIndex int, executeAtWallTime int64) {
	teamIndex := gs.Cards[sourceCardIndex].Team

	targetIndex := eff.A
	if !isValidPersistentTarget(gs, teamIndex, targetIndex) {
		return // preselected target is gone or was never set, ability fizzles
	}

	tokenCount := gs.Cards[targetIndex].Charges / 2
	tokenDefNo := eff.B

	for i := 0; i < tokenCount; i++ {
		spawnToken(gs, pq, logList, UpdateID, tokenDefNo, teamIndex, executeAtWallTime)
	}

	if gs.Cards[targetIndex].Type == state.CardTypeUnit {
		destruction.HandleUnitDestruction(gs, pq, logList, UpdateID, targetIndex, executeAtWallTime)
	} else if gs.Cards[targetIndex].Type == state.CardTypeContinuous {
		destruction.HandleContinuousDestruction(gs, pq, logList, UpdateID, targetIndex, executeAtWallTime)
	}
}

func spawnToken(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, defNo int, teamIndex int, executeAtWallTime int64) {
	def := defs.CardDefs[defNo]

	card := state.Card{
		DefNo:       defNo,
		Name:        def.Name,
		Description: def.Description,
		HP:          def.BaseHP,
		Cost:        def.Cost,
		Archetype:   def.Archetype,
		Taunt:       def.Taunt,
		Team:        teamIndex,
		Type:        def.Type,
		Zone:        state.ZoneField,
		IsToken:     def.IsToken,
		Charges:     def.Charges,
		MaxCharges:  def.MaxCharges,
	}
	for i, ability := range def.OnPlay {
		card.OnPlay[i] = state.OnPlayAbility{
			Target:      -1,
			TargetFlags: ability.TargetFlags,
		}
	}
	for i, ability := range def.Active {
		card.Actives[i] = state.ActiveAbility{
			Unlimited:   ability.Unlimited,
			Target:      -1,
			TargetFlags: ability.TargetFlags,
			Cons:        ability.Cons,
			Cost:        ability.Cost,
			CanAutoFire: ability.CanAutoFire,
			AutoAbility: false,
		}
	}

	cardIndex := len(gs.Cards)
	gs.Cards = append(gs.Cards, card)
	gs.Field[teamIndex] = append(gs.Field[teamIndex], cardIndex)

	*logList = append(*logList, logs.SpawnTokenLog{
		UpdateID:      *UpdateID,
		WallTimestamp: executeAtWallTime,
		Type:          logs.SpawnToken,
		TeamIndex:     teamIndex,
		CardIndex:     cardIndex,
		DefNo:         defNo,
		Name:          def.Name,
		Description:   def.Description,
		HP:            def.BaseHP,
		Cost:          def.Cost,
		Archetype:     def.Archetype,
		CardType:      def.Type,
		IsToken:       def.IsToken,
	})
	*UpdateID++

	schedule.ScheduleAllOnTicks(gs, pq, logList, UpdateID, cardIndex, gs.SimPauseStartedAt-gs.SimPauseOffset, executeAtWallTime)
}
