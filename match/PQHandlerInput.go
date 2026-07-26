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
	case "stage":
		match.handleStage(ie)
	case "unstage":
		match.handleUnstage(ie)
	case "autosummon":
		match.handleAutoSummon(ie)
	case "unautosummon":
		match.handleUnAutoSummon(ie)
	case "ready":
		match.handleInputReady(ie)
	case "unready":
		match.handleInputUnready(ie)
	case "active":
		match.handleUseContinuous(ie)
	case "instant":
		match.handleInstantPlay(ie)
	case "preselect":
		match.handlePreselect(ie)
	case "unitAbility":
		match.handleUnitAbility(ie)
	case "toggleAutoAbility":
		match.handleToggleAutoAbility(ie)
	case "structureDraw":
		match.handleStructureDraw(ie)
	}

}
