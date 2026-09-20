package defs

type StructureDef struct {
	Name              string
	Description       string
	BaseHP            int
	ManaRegen         int
	ManaRegenInterval int64 // milliseconds
	DrawCost          int
}

var StructureDefs = []StructureDef{
	0: {}, // empty, so DefNo lines up with index
	1: {
		Name:              "Land",
		Description:       "",
		BaseHP:            8,
		ManaRegen:         1,
		ManaRegenInterval: 1000,
		DrawCost:          0,
	},
}
