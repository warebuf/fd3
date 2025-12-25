package match

type BenchMove struct {
	Type   string `json:"Type"`
	UnitID uint   `json:"UnitID"`
}

type Manual struct {
	Type       string `json:"Type"`
	UnitID     uint   `json:"UnitID"`
	FunctionID uint   `json:"FunctionID"`
}
