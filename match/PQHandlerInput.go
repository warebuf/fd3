package match

import (
	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match/eventpq"
)

func (match *Match) handleInputEvent(ie *eventpq.InputEvent) {

	if match.GameState.Phase == state.PhaseGameOver {
		return
	}

	switch ie.Type {
	case "fullPause":
		match.handleInputPause(ie)
	case "autosummon":
		match.handleAutoSummon(ie)
	case "unautosummon":
		match.handleUnAutoSummon(ie)
	case "ready":
		match.handleInputReady(ie)
	case "unready":
		match.handleInputUnready(ie)
	case "skipChain":
		match.handleInputSkipChain(ie)
	case "unskipChain":
		match.handleInputUnskipChain(ie)
	case "instant":
		match.handleInstantPlay(ie)
	case "preselect":
		match.handlePreselect(ie)
	case "unitAbility":
		match.handleUnitAbility(ie)
	case "trapAbility":
		match.handleTrapAbility(ie)
	case "toggleAutoAbility":
		match.handleToggleAutoAbility(ie)
	case "toggleStructureAutoAbility":
		match.handleToggleStructureAutoAbility(ie)
	case "structureDraw":
		match.handleStructureDraw(ie)
	case "bid":
		match.handleBid(ie)
	}

}
