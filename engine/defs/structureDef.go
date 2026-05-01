package defs

type StructureDef struct {
	Name        string
	Description string
	BaseHP      int
}

var StructureDefs = []StructureDef{
	0: {}, // empty, so DefNo lines up with index
	1: {
		Name:        "Land",
		Description: "",
		BaseHP:      20,
	},
}
