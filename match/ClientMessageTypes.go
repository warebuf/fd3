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

type Target struct {
	Type       string `json:"Type"`
	UnitID     uint   `json:"UnitID"`
	FunctionID uint   `json:"FunctionID"`
	Target1    int    `json:"Target1"`
	Target2    int    `json:"Target2"`
}
