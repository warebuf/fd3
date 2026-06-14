package defs

type StructureDef struct {
	Name              string
	Description       string
	BaseHP            int
	ManaRegen         int
	ManaRegenInterval int64 // milliseconds
}

var StructureDefs = []StructureDef{
	0: {}, // empty, so DefNo lines up with index
	1: {
		Name:              "Land",
		Description:       "",
		BaseHP:            8,
		ManaRegen:         5,
		ManaRegenInterval: 5000,
	},
}
