package defs

type CardDef struct {
	OnPlay   []Effect
	Triggers []Trigger
}

type Effect struct {
	Kind       int
	TargetRule int
	Amount     int
}

type Trigger struct {
	Condition int
	Effect    Effect
}
