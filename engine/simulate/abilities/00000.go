package abilities

import (
	"github.com/warebuf/fd3/engine"
)

// heal 1 random on same team
type A00000 struct {

	// base info
	CasterIndex     int
	DeclarationTime int
	Window          int
	Valid           bool

	// ability level info
	Target int
	Amount int
}

func (A00000) ID() int { return 0 }

func (ability A00000) Declaration(state *engine.GameState) {

	// check if valid

	// get the random ally target

	// set the ability in pending state

}

func (ability A00000) Resolution(state *engine.GameState) {

	// check if valid, if not return

	// check if target is alive, if not return

	//

}

type S00000 struct {

	// base info
	CasterIndex int

	//

}

func (S00000) ID() int { return 0 }
