package effects

import (
	"github.com/warebuf/fd3/engine/defs"
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/sim/schedule"
	"github.com/warebuf/fd3/engine/state"
)

func spawnToken(gs *state.GameState, pq *state.SimulationPQ, logList *[]any, UpdateID *int, defNo int, teamIndex int, executeAtWallTime int64) {
	def := defs.CardDefs[defNo]

	card := state.Card{
		DefNo:          defNo,
		Name:           def.Name,
		Description:    def.Description,
		HP:             def.BaseHP,
		Cost:           def.Cost,
		Archetype:      def.Archetype,
		Taunt:          def.Taunt,
		Bodyguard:      def.Bodyguard,
		Team:           teamIndex,
		Type:           def.Type,
		Zone:           state.ZoneField,
		IsToken:        def.IsToken,
		ChainEntry:     def.ChainEntry,
		Charges:        def.Charges,
		MaxCharges:     def.MaxCharges,
		EquipHostIndex: -1,
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
		ChainEntry:    def.ChainEntry,
	})
	*UpdateID++

	schedule.ScheduleAllOnTicks(gs, pq, logList, UpdateID, cardIndex, gs.SimPauseStartedAt-gs.SimPauseOffset, executeAtWallTime)
}
