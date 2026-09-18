package state

type Structure struct {
	DefNo       int
	Name        string
	Description string
	DrawCost    int
	HP          int
	AutoAbility bool // autofires draw when affordable, player-toggleable
}
